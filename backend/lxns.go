// 落雪查分器 OAuth 接入(机密客户端跑在网关侧,浏览器只持有 session)
// 文档: https://maimai.lxns.net/docs/oauth-guide
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type lxnsConfig struct {
	clientID    string
	clientSecret string
	redirectURI string
	scopes      string
	base        string
}

func loadLxnsConfig() lxnsConfig {
	return lxnsConfig{
		clientID:     os.Getenv("LXNS_CLIENT_ID"),
		clientSecret: os.Getenv("LXNS_CLIENT_SECRET"),
		redirectURI:  getenv("LXNS_REDIRECT_URI", "http://localhost:3000/lxns-callback"),
		scopes:       getenv("LXNS_SCOPES", "openid profile read_player read_user_profile"),
		base:         getenv("LXNS_BASE", "https://maimai.lxns.net"),
	}
}

// ---------- 待授权 state(PKCE verifier 暂存,10 分钟有效) ----------

type pendingAuth struct {
	verifier  string
	createdAt time.Time
}

var pendingMu sync.Mutex
var pendingAuths = map[string]pendingAuth{}

func randB64(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func codeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// ---------- session(refresh token 落盘,access token 只放内存) ----------

type lxnsSession struct {
	ID           string `json:"id"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	Sub          string `json:"sub"`
	Name         string `json:"name"`

	accessToken string
	accessExp   time.Time
}

type lxnsStore struct {
	mu       sync.Mutex
	sessions map[string]*lxnsSession
	dataDir  string
}

func newLxnsStore(dataDir string) *lxnsStore {
	s := &lxnsStore{sessions: map[string]*lxnsSession{}, dataDir: dataDir}
	s.load()
	return s
}

func (s *lxnsStore) sessionsPath() string { return filepath.Join(s.dataDir, "lxns_sessions.json") }

func (s *lxnsStore) load() {
	b, err := os.ReadFile(s.sessionsPath())
	if err != nil {
		return
	}
	var list []lxnsSession
	if err := json.Unmarshal(b, &list); err != nil {
		return
	}
	for i := range list {
		cp := list[i]
		s.sessions[cp.ID] = &cp
	}
}

func (s *lxnsStore) save() {
	list := make([]lxnsSession, 0, len(s.sessions))
	for _, v := range s.sessions {
		list = append(list, *v)
	}
	os.MkdirAll(s.dataDir, 0o755)
	b, _ := json.Marshal(list)
	os.WriteFile(s.sessionsPath(), b, 0o600)
}

func (s *lxnsStore) get(id string) *lxnsSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessions[id]
}

func (s *lxnsStore) del(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
	s.save()
}

// ---------- 落雪 HTTP ----------

func lxnsPostToken(cfg lxnsConfig, payload map[string]string) (map[string]any, error) {
	b, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, cfg.base+"/api/v0/oauth/token", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var m map[string]any
	if err := json.Unmarshal(rb, &m); err != nil {
		return nil, fmt.Errorf("token endpoint bad response: %s", string(rb))
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%v: %v", m["error"], m["error_description"])
	}
	return m, nil
}

func lxnsAPIGet(cfg lxnsConfig, access, path string) (int, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, cfg.base+path, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+access)
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	return resp.StatusCode, rb, nil
}

// ensureAccess 保证 session 有未过期的 access token(提前 60s 刷新)
func (s *lxnsStore) ensureAccess(cfg lxnsConfig, sess *lxnsSession) error {
	s.mu.Lock()
	valid := sess.accessToken != "" && time.Now().Add(60*time.Second).Before(sess.accessExp)
	refresh := sess.RefreshToken
	s.mu.Unlock()
	if valid {
		return nil
	}
	m, err := lxnsPostToken(cfg, map[string]string{
		"client_id": cfg.clientID, "client_secret": cfg.clientSecret,
		"grant_type": "refresh_token", "refresh_token": refresh,
	})
	if err != nil {
		return err
	}
	s.mu.Lock()
	sess.accessToken, _ = m["access_token"].(string)
	exp, _ := m["expires_in"].(float64)
	sess.accessExp = time.Now().Add(time.Duration(exp) * time.Second)
	if rt, _ := m["refresh_token"].(string); rt != "" {
		sess.RefreshToken = rt // 轮换:旧 refresh 立即失效,必须保存新的
	}
	s.mu.Unlock()
	s.save()
	return nil
}

func sessionFromRequest(r *http.Request, s *lxnsStore) *lxnsSession {
	id := r.Header.Get("X-Session")
	if id == "" {
		return nil
	}
	return s.get(id)
}

// ---------- handlers ----------

func handleLxnsAuthURL(cfg lxnsConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.clientID == "" {
			writeJSON(w, 503, map[string]any{"error": "lxns not configured: set LXNS_CLIENT_ID (see backend/.env.example)"})
			return
		}
		state, err := randB64(24)
		if err != nil {
			writeJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		verifier, err := randB64(64)
		if err != nil {
			writeJSON(w, 500, map[string]any{"error": err.Error()})
			return
		}
		pendingMu.Lock()
		pendingAuths[state] = pendingAuth{verifier: verifier, createdAt: time.Now()}
		for k, v := range pendingAuths { // 顺手清理过期
			if time.Since(v.createdAt) > 10*time.Minute {
				delete(pendingAuths, k)
			}
		}
		pendingMu.Unlock()
		q := url.Values{
			"response_type": {"code"}, "client_id": {cfg.clientID},
			"redirect_uri": {cfg.redirectURI}, "scope": {cfg.scopes}, "state": {state},
			"code_challenge": {codeChallenge(verifier)}, "code_challenge_method": {"S256"},
		}
		writeJSON(w, 200, map[string]any{"auth_url": cfg.base + "/oauth/authorize?" + q.Encode()})
	}
}

func handleLxnsCallback(cfg lxnsConfig, store *lxnsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code, state := r.URL.Query().Get("code"), r.URL.Query().Get("state")
		if code == "" || state == "" {
			writeJSON(w, 400, map[string]any{"error": "missing code/state"})
			return
		}
		pendingMu.Lock()
		pa, ok := pendingAuths[state]
		if ok {
			delete(pendingAuths, state)
		}
		pendingMu.Unlock()
		if !ok {
			writeJSON(w, 400, map[string]any{"error": "invalid or expired state"})
			return
		}
		payload := map[string]string{
			"client_id": cfg.clientID, "grant_type": "authorization_code",
			"code": code, "redirect_uri": cfg.redirectURI, "code_verifier": pa.verifier,
		}
		if cfg.clientSecret != "" {
			payload["client_secret"] = cfg.clientSecret
		}
		m, err := lxnsPostToken(cfg, payload)
		if err != nil {
			writeJSON(w, 400, map[string]any{"error": "token exchange failed", "detail": err.Error()})
			return
		}
		access, _ := m["access_token"].(string)
		exp, _ := m["expires_in"].(float64)
		sid, _ := randB64(24)
		sess := &lxnsSession{ID: sid, accessToken: access, accessExp: time.Now().Add(time.Duration(exp) * time.Second)}
		if rt, _ := m["refresh_token"].(string); rt != "" {
			sess.RefreshToken = rt
		}
		if sc, _ := m["scope"].(string); sc != "" {
			sess.Scope = sc
		}
		// 拉取身份与玩家信息,登录即展示
		var userinfo, player map[string]any
		if st, rb, err := lxnsAPIGet(cfg, access, "/api/v0/oauth/userinfo"); err == nil && st == 200 {
			json.Unmarshal(rb, &userinfo)
		}
		if st, rb, err := lxnsAPIGet(cfg, access, "/api/v0/user/maimai/player"); err == nil && st == 200 {
			var env struct {
				Data map[string]any `json:"data"`
			}
			if json.Unmarshal(rb, &env) == nil {
				player = env.Data
			}
		}
		if userinfo != nil {
			sess.Sub, _ = userinfo["sub"].(string)
			sess.Name, _ = userinfo["preferred_username"].(string)
			if sess.Name == "" {
				sess.Name, _ = userinfo["name"].(string)
			}
		}
		store.mu.Lock()
		store.sessions[sid] = sess
		store.mu.Unlock()
		store.save()
		writeJSON(w, 200, map[string]any{"session": sid, "user": userinfo, "player": player})
	}
}

func handleLxnsMe(cfg lxnsConfig, store *lxnsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := sessionFromRequest(r, store)
		if sess == nil {
			writeJSON(w, 401, map[string]any{"error": "missing or invalid X-Session"})
			return
		}
		if err := store.ensureAccess(cfg, sess); err != nil {
			log.Printf("lxns refresh: %v", err)
			writeJSON(w, 401, map[string]any{"error": "session expired, please login again", "detail": err.Error()})
			return
		}
		store.mu.Lock()
		access := sess.accessToken
		store.mu.Unlock()
		var userinfo, player map[string]any
		if _, rb, err := lxnsAPIGet(cfg, access, "/api/v0/oauth/userinfo"); err == nil {
			json.Unmarshal(rb, &userinfo)
		}
		if st, rb, err := lxnsAPIGet(cfg, access, "/api/v0/user/maimai/player"); err == nil && st == 200 {
			var env struct {
				Data map[string]any `json:"data"`
			}
			if json.Unmarshal(rb, &env) == nil {
				player = env.Data
			}
		}
		writeJSON(w, 200, map[string]any{
			"user": userinfo, "player": player,
			"session": sess.ID, "scope": sess.Scope,
		})
	}
}

// 透传个人成绩(包一层 {data} 原样返回,方便前端直接用)
func handleLxnsScores(cfg lxnsConfig, store *lxnsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := sessionFromRequest(r, store)
		if sess == nil {
			writeJSON(w, 401, map[string]any{"error": "missing or invalid X-Session"})
			return
		}
		if err := store.ensureAccess(cfg, sess); err != nil {
			writeJSON(w, 401, map[string]any{"error": "session expired, please login again"})
			return
		}
		store.mu.Lock()
		access := sess.accessToken
		store.mu.Unlock()
		st, rb, err := lxnsAPIGet(cfg, access, "/api/v0/user/maimai/player/scores")
		if err != nil {
			writeJSON(w, 502, map[string]any{"error": "lxns unreachable", "detail": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(st)
		w.Write(rb)
	}
}

func handleLxnsLogout(store *lxnsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if id := r.Header.Get("X-Session"); id != "" {
			store.del(id)
		}
		writeJSON(w, 200, map[string]any{"ok": true})
	}
}
