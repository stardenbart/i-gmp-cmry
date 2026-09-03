package issueusecase

import (
	"testing"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/internal/domain/issue"
)

// fakeIssueRepo implements issue.IssueRepository with only the methods
// CheckAccess/GetByID actually touch doing real work.
type fakeIssueRepo struct {
	byID       map[string]*issue.Issue
	plantByID  map[string]*string
	scopedUser map[string]string // issueID -> the one userID IsScopedToUser allows
}

func (r *fakeIssueRepo) FindAll(int, int, string, string, string, *bool) ([]issue.Issue, int64, error) {
	return nil, 0, nil
}
func (r *fakeIssueRepo) FindByID(id string) (*issue.Issue, error) {
	item, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return item, nil
}
func (r *fakeIssueRepo) FindByResultID(string) (*issue.Issue, error) { return nil, nil }
func (r *fakeIssueRepo) FindActiveByUraianAndDetailKawasan(string, string) (*issue.Issue, error) {
	return nil, nil
}
func (r *fakeIssueRepo) FindActiveByResultContext(string) (*issue.Issue, error) { return nil, nil }
func (r *fakeIssueRepo) IsScopedToUser(issueID, userID string) (bool, error) {
	want, ok := r.scopedUser[issueID]
	return ok && want == userID, nil
}
func (r *fakeIssueRepo) FindPlantID(issueID string) (*string, error) {
	return r.plantByID[issueID], nil
}
func (r *fakeIssueRepo) FindReminderCandidates() ([]issue.Issue, error) { return nil, nil }
func (r *fakeIssueRepo) ConsolidateDuplicateActiveIssues() error        { return nil }
func (r *fakeIssueRepo) Create(*issue.Issue) error                      { return nil }
func (r *fakeIssueRepo) Update(*issue.Issue) error                      { return nil }
func (r *fakeIssueRepo) Delete(string) error                            { return nil }

// fakeAuthUserRepo implements authdomain.UserRepository — GetByID looks up
// the issue's PIC display name through this.
type fakeAuthUserRepo struct{}

func (fakeAuthUserRepo) FindAll(int, int, string, string, string, string) ([]authdomain.User, int64, error) {
	return nil, 0, nil
}
func (fakeAuthUserRepo) FindByID(string) (*authdomain.User, error)       { return nil, nil }
func (fakeAuthUserRepo) FindByUsername(string) (*authdomain.User, error) { return nil, nil }
func (fakeAuthUserRepo) FindByEmail(string) (*authdomain.User, error)    { return nil, nil }
func (fakeAuthUserRepo) Create(*authdomain.User) error                   { return nil }
func (fakeAuthUserRepo) Update(*authdomain.User) error                   { return nil }
func (fakeAuthUserRepo) Delete(string) error                             { return nil }

func testPlantID(s string) *string { return &s }

func newTestIssueUseCase(repo *fakeIssueRepo) issue.IssueUseCase {
	return NewIssueUseCase(repo, nil, nil, nil, nil, nil, fakeAuthUserRepo{}, nil, nil, nil, nil, nil, nil)
}

func TestCheckAccessDeniesCrossPlantCaller(t *testing.T) {
	repo := &fakeIssueRepo{
		byID:      map[string]*issue.Issue{"ISS-1": {IssueID: "ISS-1", IssuePICUserID: "USR-1"}},
		plantByID: map[string]*string{"ISS-1": testPlantID("PLT-A")},
	}
	uc := newTestIssueUseCase(repo)

	if _, err := uc.CheckAccess("ISS-1", "PLT-B", "USR-1", true); err == nil {
		t.Fatal("a caller scoped to a different plant must not see this issue, even as an auditor")
	}
}

func TestCheckAccessDeniesAuditeeNotAssigned(t *testing.T) {
	repo := &fakeIssueRepo{
		byID:       map[string]*issue.Issue{"ISS-1": {IssueID: "ISS-1", IssuePICUserID: "USR-1"}},
		plantByID:  map[string]*string{"ISS-1": testPlantID("PLT-A")},
		scopedUser: map[string]string{"ISS-1": "USR-1"},
	}
	uc := newTestIssueUseCase(repo)

	if _, err := uc.CheckAccess("ISS-1", "PLT-A", "USR-2", false); err == nil {
		t.Fatal("an auditee not assigned to this issue must not be able to see it — this is the bug where one auditee could read another auditee's findings")
	}
}

func TestCheckAccessAllowsAssignedAuditee(t *testing.T) {
	repo := &fakeIssueRepo{
		byID:       map[string]*issue.Issue{"ISS-1": {IssueID: "ISS-1", IssuePICUserID: "USR-1"}},
		plantByID:  map[string]*string{"ISS-1": testPlantID("PLT-A")},
		scopedUser: map[string]string{"ISS-1": "USR-1"},
	}
	uc := newTestIssueUseCase(repo)

	item, err := uc.CheckAccess("ISS-1", "PLT-A", "USR-1", false)
	if err != nil {
		t.Fatalf("CheckAccess() error = %v", err)
	}
	if item.IssueID != "ISS-1" {
		t.Fatal("expected the assigned issue to be returned")
	}
}

func TestCheckAccessAllowsAuditorRegardlessOfPICAssignment(t *testing.T) {
	repo := &fakeIssueRepo{
		byID:      map[string]*issue.Issue{"ISS-1": {IssueID: "ISS-1", IssuePICUserID: "USR-1"}},
		plantByID: map[string]*string{"ISS-1": testPlantID("PLT-A")},
		// Deliberately no scopedUser entry — an auditor must not need it.
	}
	uc := newTestIssueUseCase(repo)

	if _, err := uc.CheckAccess("ISS-1", "PLT-A", "USR-AUDITOR", true); err != nil {
		t.Fatalf("an auditor in the right plant must see any issue regardless of PIC assignment, got error: %v", err)
	}
}

func TestCheckAccessAllowsUnscopedCaller(t *testing.T) {
	repo := &fakeIssueRepo{
		byID:      map[string]*issue.Issue{"ISS-1": {IssueID: "ISS-1", IssuePICUserID: "USR-1"}},
		plantByID: map[string]*string{"ISS-1": testPlantID("PLT-A")},
	}
	uc := newTestIssueUseCase(repo)

	if _, err := uc.CheckAccess("ISS-1", "", "USR-SUPERADMIN", true); err != nil {
		t.Fatalf("an unscoped (Super Admin / global) caller must see any issue, got error: %v", err)
	}
}
