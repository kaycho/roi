package roi

import (
	"reflect"
	"testing"
	"time"
)

var testShotA = &Shot{
	Project:       testProject.Project,
	Shot:          "CG_0010",
	Status:        ShotInProgress,
	EditOrder:     10,
	Description:   "방에 우두커니 혼자 않아 있는 로이.",
	CGDescription: "조명판 들고 있는 사람이 촬영되었으니 지워주세요.",
	TimecodeIn:    "00:00:00:01",
	TimecodeOut:   "00:00:05:12",
	Duration:      132,
	Tags:          []string{"로이", "리무브"},
	WorkingTasks:  []string{"fx_fire"}, // testTaskA 확인
}

var testShotB = &Shot{
	Shot:          "CG_0020",
	Project:       testProject.Project,
	Status:        ShotWaiting,
	EditOrder:     20,
	Description:   "고개를 돌려 창문 밖을 바라본다.",
	CGDescription: "전반적인 느낌을 어둡게 바꿔주세요.",
	TimecodeIn:    "00:00:05:12",
	TimecodeOut:   "00:00:06:03",
	Duration:      15,
	Tags:          []string{"로이", "창문"},
}
var testShotC = &Shot{
	Shot:          "CG_0030",
	Project:       testProject.Project,
	Status:        ShotWaiting,
	EditOrder:     30,
	Description:   "쓸쓸해 보이는 가로등",
	CGDescription: "가로등이 너무 깨끗하니 레트로 한 느낌을 살려주세요.",
	TimecodeIn:    "00:00:06:03",
	TimecodeOut:   "00:00:08:15",
	Duration:      36,
	Tags:          []string{"가로등", "창문"},
}

var testShots = []*Shot{testShotA, testShotB, testShotC}

func TestShot(t *testing.T) {
	want := testShots

	db, err := testDB()
	if err != nil {
		t.Fatalf("could not connect to database: %v", err)
	}
	err = AddProject(db, testProject)
	if err != nil {
		t.Fatalf("could not add project to projects table: %s", err)
	}

	for _, s := range want {
		err = AddShot(db, testProject.Project, s)
		if err != nil {
			t.Fatalf("could not add shot to shots table: %s", err)
		}
		exist, err := ShotExist(db, testProject.Project, s.Shot)
		if err != nil {
			t.Fatalf("could not check shot existence from shots table: %s", err)
		}
		if !exist {
			t.Fatalf("shot not found from shots table: %s", s.Shot)
		}
		got, err := GetShot(db, testProject.Project, s.Shot)
		if err != nil {
			t.Fatalf("could not get shot from shots table: %s", err)
		}
		if !IsValidShot(got.Shot) {
			if err != nil {
				t.Fatalf("find shot with invalid id from shots table: %s", err)
			}
		}
		if !reflect.DeepEqual(got, s) {
			t.Fatalf("got: %v, want: %v", got, s)
		}
	}

	got, err := SearchShots(db, testProject.Project, "", "", "", "", "", time.Time{})
	if err != nil {
		t.Fatalf("could not search shots from shots table: %s", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v, want: %v", got, want)
	}

	got, err = SearchShots(db, testProject.Project, "CG_0010", "", "", "", "", time.Time{})
	if err != nil {
		t.Fatalf("could not search shots from shots table: %s", err)
	}
	want = []*Shot{testShotA}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v, want: %v", got, want)
	}
	got, err = SearchShots(db, testProject.Project, "", "로이", "", "", "", time.Time{})
	if err != nil {
		t.Fatalf("could not search shots from shots table: %s", err)
	}
	want = []*Shot{testShotA, testShotB}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v, want: %v", got, want)
	}

	for _, s := range want {
		err = UpdateShot(db, testProject.Project, s.Shot, UpdateShotParam{Status: ShotWaiting})
		if err != nil {
			t.Fatalf("could not clear(update) shot: %s", err)
		}
		err = DeleteShot(db, testProject.Project, s.Shot)
		if err != nil {
			t.Fatalf("could not delete shot from shots table: %s", err)
		}
	}

	err = DeleteProject(db, testProject.Project)
	if err != nil {
		t.Fatalf("could not delete project: %s", err)
	}
}
