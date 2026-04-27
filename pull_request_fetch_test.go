package main

import (
	"testing"
)

func TestParsePullRequestListJSON_Array(t *testing.T) {
	raw := `[{"id":46573783,"iid":165,"title":"修改trafficservice前缀","source_branch":"traffic_service","target_branch":"master","state":"opened","created_at":"2026-04-20T16:16:32.501+08:00","updated_at":"2026-04-20T16:16:34.964+08:00","web_url":"https://example.com/mr/165","author":{"name":"[姓名]","username":"u1","name_cn":"[中文名]"}}]`
	got, err := parsePullRequestListJSON([]byte(raw), 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("len items=%d", len(got.Items))
	}
	it := got.Items[0]
	if it.ID != 46573783 || it.Number != 165 || it.Title != "修改trafficservice前缀" {
		t.Fatalf("%+v", it)
	}
	if it.URL != "https://example.com/mr/165" {
		t.Fatal(it.URL)
	}
	if it.SourceBranch != "traffic_service" || it.TargetBranch != "master" {
		t.Fatalf("%+v", it)
	}
	if it.State != "open" {
		t.Fatal(it.State)
	}
	if it.Author != "[中文名]" {
		t.Fatal(it.Author)
	}
	if got.Page != 1 || got.PageSize != 10 || got.TotalPages != 1 || got.Total != 1 {
		t.Fatalf("pagination %+v", got)
	}
}

func TestParsePullRequestListJSON_LegacyEnvelope(t *testing.T) {
	raw := `{"items":[{"id":1,"number":2,"url":"http://x","title":"t","author":"a","sourceBranch":"s","targetBranch":"m","state":"merged","createdAt":"c","updatedAt":"u"}],"total":1,"page":1,"pageSize":10,"totalPages":1}`
	got, err := parsePullRequestListJSON([]byte(raw), 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 || got.Items[0].Number != 2 || got.Total != 1 {
		t.Fatalf("%+v", got)
	}
}

func TestDerivePullRequestPagination(t *testing.T) {
	tcases := []struct {
		n, page, ps, wantTot, wantTP int
	}{
		{0, 1, 10, 0, 1},
		{0, 3, 10, 10, 2},
		{3, 2, 10, 13, 2},
		{10, 1, 10, 20, 2},
	}
	for _, tc := range tcases {
		tot, tp := derivePullRequestPagination(tc.n, tc.page, tc.ps)
		if tot != tc.wantTot || tp != tc.wantTP {
			t.Fatalf("n=%d page=%d ps=%d got total=%d tp=%d want %d %d", tc.n, tc.page, tc.ps, tot, tp, tc.wantTot, tc.wantTP)
		}
	}
}

func TestParsePullRequestTotalResponse(t *testing.T) {
	for _, tc := range []struct {
		raw string
		n   int
	}{
		{`47`, 47},
		{`{"total": 47}`, 47},
		{`{"data": 12}`, 12},
		{`{"data":{"total":5}}`, 5},
	} {
		got, err := parsePullRequestTotalResponse([]byte(tc.raw))
		if err != nil || got != tc.n {
			t.Fatalf("%q: got %d err %v want %d", tc.raw, got, err, tc.n)
		}
	}
}

func TestExtractPullRequestTotalFromListBody(t *testing.T) {
	tot, def, err := extractPullRequestTotalFromListBody([]byte(`{"total":99,"items":[]}`))
	if err != nil || !def || tot != 99 {
		t.Fatalf("got %d def %v err %v", tot, def, err)
	}
	_, def, err = extractPullRequestTotalFromListBody([]byte(`[]`))
	if err != nil || def {
		t.Fatalf("array: def %v err %v", def, err)
	}
	_, def, err = extractPullRequestTotalFromListBody([]byte(`{"items":[]}`))
	if err != nil || def {
		t.Fatalf("no total: def %v err %v", def, err)
	}
}

func TestCountPullRequestItemsInListBody(t *testing.T) {
	n, err := countPullRequestItemsInListBody([]byte(`[{},{}]`))
	if err != nil || n != 2 {
		t.Fatalf("got %d err %v", n, err)
	}
	n, err = countPullRequestItemsInListBody([]byte(`{"items":[{},{}]}`))
	if err != nil || n != 2 {
		t.Fatalf("envelope got %d err %v", n, err)
	}
}
