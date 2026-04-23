package clone

import (
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/githubclient"
	"github.com/DevViking-Persike/njord-cli/internal/gitlabclient"
)

func TestFromGitLab(t *testing.T) {
	in := gitlabclient.ProjectInfo{
		PathWithNamespace: "grp/repo",
		Description:       "api",
		SSHURLToRepo:      "git@gitlab.com:grp/repo.git",
		WebURL:            "https://gitlab.com/grp/repo",
	}
	got := FromGitLab(in)
	if got.FullName != "grp/repo" || got.CloneSSH != in.SSHURLToRepo || got.Host != HostGitLab {
		t.Fatalf("FromGitLab mapping wrong: %+v", got)
	}
}

func TestFromGitHub(t *testing.T) {
	in := githubclient.Repo{
		FullName:    "owner/repo",
		Description: "lib",
		SSHURL:      "git@github.com:owner/repo.git",
		HTMLURL:     "https://github.com/owner/repo",
	}
	got := FromGitHub(in)
	if got.FullName != "owner/repo" || got.CloneSSH != in.SSHURL || got.Host != HostGitHub {
		t.Fatalf("FromGitHub mapping wrong: %+v", got)
	}
}

func TestFilterRepos(t *testing.T) {
	list := []Repo{
		{FullName: "grp/api-gateway", Description: "API do gateway", Host: HostGitLab},
		{FullName: "grp/worker", Description: "job queue worker", Host: HostGitLab},
		{FullName: "me/dotfiles", Description: "zsh config", Host: HostGitHub},
	}

	cases := []struct {
		name  string
		query string
		want  []string
	}{
		{"empty query keeps all", "", []string{"grp/api-gateway", "grp/worker", "me/dotfiles"}},
		{"case insensitive name", "WORKER", []string{"grp/worker"}},
		{"match description", "gateway", []string{"grp/api-gateway"}},
		{"multiple tokens AND", "grp api", []string{"grp/api-gateway"}},
		{"no match", "nothing-here", []string{}},
		{"trim spaces", "  api  ", []string{"grp/api-gateway"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FilterRepos(list, tc.query)
			if len(got) != len(tc.want) {
				t.Fatalf("len = %d (%v), want %v", len(got), names(got), tc.want)
			}
			for i, r := range got {
				if r.FullName != tc.want[i] {
					t.Fatalf("[%d] = %q, want %q", i, r.FullName, tc.want[i])
				}
			}
		})
	}
}

func TestSortByName(t *testing.T) {
	list := []Repo{
		{FullName: "zeta/y"},
		{FullName: "Alpha/x"},
		{FullName: "beta/z"},
	}
	SortByName(list)
	want := []string{"Alpha/x", "beta/z", "zeta/y"}
	for i, r := range list {
		if r.FullName != want[i] {
			t.Fatalf("[%d] = %q, want %q", i, r.FullName, want[i])
		}
	}
}

func TestGroupFromGitLab(t *testing.T) {
	in := gitlabclient.GroupInfo{ID: 7, Name: "bill", FullPath: "avitaseg/bill", FullName: "avitaseg / bill"}
	got := GroupFromGitLab(in)
	if got.ID != 7 || got.FullPath != "avitaseg/bill" || got.FullName != "avitaseg / bill" {
		t.Fatalf("mapping wrong: %+v", got)
	}
}

func TestCollapseToTopBuckets(t *testing.T) {
	t.Run("collapse subgroups under shared prefix", func(t *testing.T) {
		in := []Group{
			{FullPath: "avitaseg"},
			{FullPath: "avitaseg/bill"},
			{FullPath: "avitaseg/bill/bibliotecas"},
			{FullPath: "avitaseg/bill/bibliotecas/angular"},
			{FullPath: "avitaseg/gap"},
			{FullPath: "avitaseg/jnd"},
			{FullPath: "avitaseg/jnd/bibliotecas"},
			{FullPath: "avitaseg/pap"},
			{FullPath: "avitaseg/pap/plataforma"},
			{FullPath: "avitaseg/sie"},
		}
		got := CollapseToTopBuckets(in)
		want := []string{"avitaseg/bill", "avitaseg/gap", "avitaseg/jnd", "avitaseg/pap", "avitaseg/sie"}
		if len(got) != len(want) {
			t.Fatalf("len = %d (%v), want %v", len(got), pathsOf(got), want)
		}
		for i, g := range got {
			if g.FullPath != want[i] {
				t.Fatalf("[%d] = %q, want %q", i, g.FullPath, want[i])
			}
		}
	})

	t.Run("no common prefix → top-level namespaces", func(t *testing.T) {
		in := []Group{
			{FullPath: "orga"},
			{FullPath: "orga/x"},
			{FullPath: "orgb"},
			{FullPath: "orgb/y"},
		}
		got := CollapseToTopBuckets(in)
		want := []string{"orga", "orgb"}
		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", pathsOf(got), want)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		if got := CollapseToTopBuckets(nil); len(got) != 0 {
			t.Fatalf("expected empty, got %v", pathsOf(got))
		}
	})

	t.Run("single root falls back to original", func(t *testing.T) {
		in := []Group{{FullPath: "avitaseg"}}
		got := CollapseToTopBuckets(in)
		if len(got) != 1 || got[0].FullPath != "avitaseg" {
			t.Fatalf("expected fallback to original, got %v", pathsOf(got))
		}
	})
}

func pathsOf(gs []Group) []string {
	out := make([]string, 0, len(gs))
	for _, g := range gs {
		out = append(out, g.FullPath)
	}
	return out
}

func TestFilterGroups(t *testing.T) {
	list := []Group{
		{FullPath: "avitaseg/bill", FullName: "avitaseg / bill"},
		{FullPath: "avitaseg/gap", FullName: "avitaseg / gap"},
		{FullPath: "avitaseg/pap/plataforma", FullName: "avitaseg / pap / plataforma"},
	}
	cases := []struct {
		query string
		want  []string
	}{
		{"", []string{"avitaseg/bill", "avitaseg/gap", "avitaseg/pap/plataforma"}},
		{"BILL", []string{"avitaseg/bill"}},
		{"pap plataforma", []string{"avitaseg/pap/plataforma"}},
		{"missing", []string{}},
	}
	for _, tc := range cases {
		got := FilterGroups(list, tc.query)
		if len(got) != len(tc.want) {
			t.Fatalf("query=%q len=%d want %v", tc.query, len(got), tc.want)
		}
		for i, g := range got {
			if g.FullPath != tc.want[i] {
				t.Fatalf("query=%q [%d]=%q want %q", tc.query, i, g.FullPath, tc.want[i])
			}
		}
	}
}

func names(rs []Repo) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.FullName)
	}
	return out
}
