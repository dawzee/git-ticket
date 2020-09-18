package tests

import (
	"testing"

	"github.com/daedaleanai/git-ticket/bug"
	"github.com/daedaleanai/git-ticket/misc/random_bugs"
	"github.com/daedaleanai/git-ticket/repository"
)

func TestReadBugs(t *testing.T) {
	repo := repository.CreateTestRepo(false)
	defer repository.CleanupTestRepos(repo)

	random_bugs.FillRepoWithSeed(repo, 15, 42)

	bugs := bug.ReadAllLocalBugs(repo)
	for b := range bugs {
		if b.Err != nil {
			t.Fatal(b.Err)
		}
	}
}

func benchmarkReadBugs(bugNumber int, t *testing.B) {
	repo := repository.CreateTestRepo(false)
	defer repository.CleanupTestRepos(repo)

	random_bugs.FillRepoWithSeed(repo, bugNumber, 42)
	t.ResetTimer()

	for n := 0; n < t.N; n++ {
		bugs := bug.ReadAllLocalBugs(repo)
		for b := range bugs {
			if b.Err != nil {
				t.Fatal(b.Err)
			}
		}
	}
}

func BenchmarkReadBugs5(b *testing.B)   { benchmarkReadBugs(5, b) }
func BenchmarkReadBugs25(b *testing.B)  { benchmarkReadBugs(25, b) }
func BenchmarkReadBugs150(b *testing.B) { benchmarkReadBugs(150, b) }
