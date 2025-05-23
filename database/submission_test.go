package database

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"
)

func TestSubmission(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	if err := RegisterUser(db, "user1", "id1"); err != nil {
		t.Fatal(err)
	}

	sub := Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
		Source:      "source",
	}

	id, err := SaveSubmission(db, sub)
	if err != nil {
		t.Fatal(err)
	}

	sub2, err := FetchSubmission(db, id)

	if err != nil {
		t.Fatal(err)
	}

	if sub2.User.Name != "user1" || sub2.Problem.Name != "aplusb" {
		t.Fatal("invalid data", sub2)
	}
}

func TestFetchInvalidSubmission(t *testing.T) {
	db := CreateTestDB(t)

	_, err := FetchSubmission(db, 123)

	if err != ErrNotExist {
		t.Fatal(err)
	}
}

func TestUpdateSubmissionStatus(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	if err := RegisterUser(db, "user1", "id1"); err != nil {
		t.Fatal(err)
	}

	sub := Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
		MaxTime:     1234,
		Source:      "source",
	}

	id, err := SaveSubmission(db, sub)
	if err != nil {
		t.Fatal(err)
	}
	if err := UpdateSubmissionStatus(db, id, "IE"); err != nil {
		t.Fatal(err)
	}

	sub2, err := FetchSubmission(db, id)

	if err != nil {
		t.Fatal(err)
	}

	if sub2.User.Name != "user1" || sub2.Problem.Name != "aplusb" || sub2.Status != "IE" || sub2.MaxTime != 1234 {
		t.Fatal("invalid data", sub2)
	}
}

func TestSubmitInvalidSource(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	if err := RegisterUser(db, "user1", "id1"); err != nil {
		t.Fatal(err)
	}

	if _, err := SaveSubmission(db, Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: false},
		Source:      "",
	}); err == nil {
		t.Fatal("must be error")
	}

	if _, err := SaveSubmission(db, Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: false},
		Source:      strings.Repeat("a", 1024*1024+1),
	}); err == nil {
		t.Fatal("must be error")
	}

}

func TestSubmissionResult(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	if err := RegisterUser(db, "user1", "id1"); err != nil {
		t.Fatal(err)
	}

	sub := Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
		Source:      "source",
	}

	id, err := SaveSubmission(db, sub)
	if err != nil {
		t.Fatal(err)
	}

	result := SubmissionTestcaseResult{
		Submission: id,
		Testcase:   "case1.in",
		Status:     "AC",
		Time:       123,
		Memory:     456,
		Stderr:     []byte{12, 34},
	}
	if err := SaveTestcaseResults(db, []SubmissionTestcaseResult{result}); err != nil {
		t.Fatal(err)
	}

	actual, err := FetchTestcaseResults(db, id)
	if err != nil {
		t.Fatal(err)
	}

	if len(actual) != 1 || !reflect.DeepEqual(actual[0], result) {
		t.Fatal(actual, "!=", result)
	}
}

func TestSubmissionResultEmpty(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	if err := RegisterUser(db, "user1", "id1"); err != nil {
		t.Fatal(err)
	}

	sub := Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
		Source:      "source",
	}
	if _, err := SaveSubmission(db, sub); err != nil {
		t.Fatal(err)
	}

	actual, err := FetchTestcaseResults(db, sub.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(actual) != 0 {
		t.Fatal(actual, "is not empty")
	}
}

func TestSubmissionList(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	if err := RegisterUser(db, "user1", "id1"); err != nil {
		t.Fatal(err)
	}

	if _, err := SaveSubmission(db, Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
		MaxTime:     1234,
		Source:      "source",
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := SaveSubmission(db, Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: false},
		MaxTime:     123,
		Source:      "source",
	}); err != nil {
		t.Fatal(err)
	}

	{
		subs, count, err := FetchSubmissionList(db, "", "", "", "", false, []SubmissionOrder{ID_DESC}, 0, 1)

		if err != nil {
			t.Fatal(err)
		}

		if count != 2 {
			t.Fatal("count is not 2: ", count)
		}

		if len(subs) != 1 {
			t.Fatal("len(subs) is not 1: ", len(subs))
		}
	}
	{
		// problem filter
		subs, count, err := FetchSubmissionList(db, "aplusb", "", "", "", false, []SubmissionOrder{ID_DESC}, 0, 1)

		if err != nil {
			t.Fatal(err)
		}

		if count != 2 {
			t.Fatal("count is not 2: ", count)
		}

		if len(subs) != 1 {
			t.Fatal("len(subs) is not 1: ", len(subs))
		}
		if subs[0].Problem.Name != "aplusb" {
			t.Fatal("subs[0].Problem.Name is not aplusb: ", subs[0])
		}
	}
	{
		// invalid problem filter
		subs, count, err := FetchSubmissionList(db, "aplusb-dummy", "", "", "", false, []SubmissionOrder{ID_DESC}, 0, 1)

		if err != nil {
			t.Fatal(err)
		}

		if count != 0 {
			t.Fatal("count is not 0: ", count)
		}

		if len(subs) != 0 {
			t.Fatal("len(subs) is not 0: ", len(subs))
		}
	}
	{
		// sort
		subs, count, err := FetchSubmissionList(db, "", "", "", "", false, []SubmissionOrder{MAX_TIME_ASC}, 0, 1)

		if err != nil {
			t.Fatal(err)
		}

		if count != 2 {
			t.Fatal("count is not 2: ", count)
		}

		if len(subs) != 1 {
			t.Fatal("len(subs) is not : ", len(subs))
		}
		if subs[0].MaxTime != 123 {
			t.Fatal("subs[0].MaxTime is not 123: ", subs[0])
		}
	}
}

func TestDedupSubmissionList(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	if err := RegisterUser(db, "user1", "id1"); err != nil {
		t.Fatal(err)
	}
	if err := RegisterUser(db, "user2", "id2"); err != nil {
		t.Fatal(err)
	}

	if _, err := SaveSubmission(db, Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
		MaxTime:     123,
		Source:      "source",
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := SaveSubmission(db, Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
		MaxTime:     1234,
		Source:      "source",
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := SaveSubmission(db, Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user2"},
		MaxTime:     234,
		Source:      "source",
	}); err != nil {
		t.Fatal(err)
	}

	{
		subs, count, err := FetchSubmissionList(db, "", "", "", "", true, []SubmissionOrder{ID_DESC}, 0, 1)

		if err != nil {
			t.Fatal(err)
		}

		if count != 2 {
			t.Fatal("count is not 2: ", count)
		}

		if len(subs) != 1 {
			t.Fatal("len(subs) is not 1: ", len(subs))
		}

		if subs[0].UserName.String != "user2" {
			t.Fatal("subs[0].UserName is not user2: ", subs[0])
		}
	}

	{
		subs, count, err := FetchSubmissionList(db, "", "", "", "", true, []SubmissionOrder{MAX_TIME_ASC, ID_DESC}, 0, 1)

		if err != nil {
			t.Fatal(err)
		}

		if count != 2 {
			t.Fatal("count is not 2: ", count)
		}

		if len(subs) != 1 {
			t.Fatal("len(subs) is not 1: ", len(subs))
		}

		if subs[0].MaxTime != 123 {
			t.Fatal("subs[0].MaxTime is not 123: ", subs[0])
		}
	}
}
