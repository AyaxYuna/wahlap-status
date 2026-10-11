package main

import (
	"encoding/json"
	"testing"
)

// 构造:现版本 25500(2 曲),旧版本 20000(2 曲),宴会场 1 曲应被剔除
func TestComputeB50(t *testing.T) {
	scores := `{"success":true,"code":200,"data":[
		{"id":1,"song_name":"new-a","type":"dx","level_index":4,"achievements":100.5,"dx_rating":300.7},
		{"id":2,"song_name":"new-b","type":"standard","level_index":3,"achievements":99.0,"dx_rating":200.2},
		{"id":3,"song_name":"old-a","type":"standard","level_index":4,"achievements":98.0,"dx_rating":250.9},
		{"id":4,"song_name":"old-b","type":"dx","level_index":2,"achievements":97.0,"dx_rating":100.1},
		{"id":100001,"song_name":"utage","type":"utage","level_index":0,"achievements":100.0,"dx_rating":999.0}
	]}`
	songs := `{"songs":[
		{"id":1,"version":25500},{"id":2,"version":25500},
		{"id":3,"version":20000},{"id":4,"version":20000}
	],"versions":[{"version":20000},{"version":25500}]}`

	out, err := computeB50([]byte(scores), []byte(songs))
	if err != nil {
		t.Fatalf("computeB50: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	dx := m["dx"].([]any)
	std := m["standard"].([]any)
	if len(dx) != 2 || len(std) != 2 {
		t.Fatalf("got dx=%d std=%d, want 2/2", len(dx), len(std))
	}
	// 排序:各组第一应为组内最高
	if dx[0].(map[string]any)["song_name"] != "new-a" {
		t.Fatalf("dx order wrong: %v", dx[0])
	}
	if std[0].(map[string]any)["song_name"] != "old-a" {
		t.Fatalf("std order wrong: %v", std[0])
	}
	// 总分向下取整求和
	if m["dx_total"] != float64(300+200) {
		t.Fatalf("dx_total=%v want 500", m["dx_total"])
	}
	if m["standard_total"] != float64(250+100) {
		t.Fatalf("standard_total=%v want 350", m["standard_total"])
	}
}
