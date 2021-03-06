package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/studio2l/roi"
)

// apiOK는 api 질의가 잘 처리되었을 때
// 그 응답을 roi.APIResponse.Msg에 담아 반환한다.
func apiOK(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusOK)
	resp, _ := json.Marshal(roi.APIResponse{Msg: msg})
	w.Write(resp)
}

// apiInternalServerError는 내부적인 이유로 api 질의의 처리에 실패했을 때
// 이를 질의자에게 알린다. (이유는 알리지 않는다.)
func apiInternalServerError(w http.ResponseWriter) {
	resp, _ := json.Marshal(roi.APIResponse{Err: "internal error"})
	http.Error(w, string(resp), http.StatusInternalServerError)
}

// apiBadRequest는 api 질의에 문제가 있었을 때
// 그 문제를 apiReponse.Err에 담아 반환한다.
func apiBadRequest(w http.ResponseWriter, err error) {
	if err == nil {
		// err 가 nil이어서는 안되지만, 패닉을 일으키는 것보다는 낫다.
		err = errors.New("error not explained")
	}
	resp, _ := json.Marshal(roi.APIResponse{Err: err.Error()})
	http.Error(w, string(resp), http.StatusInternalServerError)
}

// addProjectApiHander는 사용자가 api를 통해 프로젝트를 생성할수 있도록 한다.
// 결과는 roi.APIResponse의 json 형식으로 반환된다.
func addProjectApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to db: %v", err)
		apiInternalServerError(w)
		return
	}

	prj := r.PostFormValue("project")
	if prj == "" {
		apiBadRequest(w, fmt.Errorf("'id' not specified"))
		return
	}
	exist, err := roi.ProjectExist(db, prj)
	if err != nil {
		log.Printf("could not check project %q exist: %v", prj, err)
		apiInternalServerError(w)
		return
	}
	if exist {
		apiBadRequest(w, fmt.Errorf("project '%s' already exists", prj))
		return
	}
	tasks := fields(r.Form.Get("default_tasks"), ",")
	p := &roi.Project{
		Project:      prj,
		DefaultTasks: tasks,
	}
	err = roi.AddProject(db, p)
	if err != nil {
		log.Printf("could not add project: %v", err)
		apiInternalServerError(w)
		return
	}
	apiOK(w, fmt.Sprintf("successfully add a project: '%s'", prj))
}

// addShotApiHander는 사용자가 api를 통해 샷을 생성할수 있도록 한다.
// 결과는 roi.APIResponse의 json 형식으로 반환된다.
func addShotApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to db: %v", err)
		apiInternalServerError(w)
		return
	}

	prj := r.PostFormValue("project")
	if prj == "" {
		apiBadRequest(w, fmt.Errorf("'project' not specified"))
		return
	}
	exist, err := roi.ProjectExist(db, prj)
	if err != nil {
		log.Printf("could not check project %q exist: %v", prj, err)
		apiInternalServerError(w)
		return
	}
	if !exist {
		apiBadRequest(w, fmt.Errorf("project '%s' not exists", prj))
		return
	}

	shot := r.PostFormValue("shot")
	if shot == "" {
		apiBadRequest(w, fmt.Errorf("'shot' not specified"))
		return
	}
	if !roi.IsValidShot(shot) {
		apiBadRequest(w, fmt.Errorf("shot id '%s' is not valid", shot))
		return
	}
	exist, err = roi.ShotExist(db, prj, shot)
	if err != nil {
		log.Printf("could not check shot '%s' exist: %v", shot, err)
		apiInternalServerError(w)
		return
	}
	if exist {
		apiBadRequest(w, fmt.Errorf("shot '%s' already exists", shot))
		return
	}

	status := "waiting"
	v := r.PostFormValue("status")
	if v != "" {
		status = v
	}
	editOrder := 0
	v = r.PostFormValue("edit_order")
	if v != "" {
		// strconv.Itoa 대신 ParseFloat을 쓰는 이유는 엑셀에서 분명히 정수형으로
		// 입력한 데이터를 excelize 패키지를 통해 받을 때 실수형으로 변경되기 때문이다.
		// 무엇이 문제인지는 확실치 않지만, 실수형으로 받은후 정수로 변환하는 것이
		// 안전하다.
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			apiBadRequest(w, fmt.Errorf("could not convert duration to int: %s", v))
			return
		}
		editOrder = int(f)
	}
	duration := 0
	v = r.PostFormValue("duration")
	if v != "" {
		// 위 내용 참조
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			apiBadRequest(w, fmt.Errorf("could not convert duration to int: %s", v))
			return
		}
		duration = int(f)
	}
	tasks := fields(r.Form.Get("working_tasks"), ",")
	if len(tasks) == 0 {
		p, err := roi.GetProject(db, prj)
		if err != nil {
			log.Printf("could not get project: %v", err)
			apiInternalServerError(w)
			return
		}
		tasks = p.DefaultTasks
	}
	s := &roi.Shot{
		Project:       prj,
		Shot:          shot,
		Status:        roi.ShotStatus(status),
		EditOrder:     editOrder,
		Description:   r.PostFormValue("description"),
		CGDescription: r.PostFormValue("cg_description"),
		TimecodeIn:    r.PostFormValue("timecode_in"),
		TimecodeOut:   r.PostFormValue("timecode_out"),
		Duration:      duration,
		Tags:          strings.Split(r.PostFormValue("tags"), ","),
		WorkingTasks:  tasks,
	}
	err = roi.AddShot(db, prj, s)
	if err != nil {
		log.Printf("could not add shot: %v", err)
		apiInternalServerError(w)
		return
	}
	for _, task := range tasks {
		t := &roi.Task{
			Project: prj,
			Shot:    shot,
			Task:    task,
			Status:  roi.TaskNotSet,
			DueDate: time.Time{},
		}
		err := roi.AddTask(db, prj, shot, t)
		if err != nil {
			log.Printf("could not add task for shot: %v", err)
			apiInternalServerError(w)
			return
		}
	}
	apiOK(w, fmt.Sprintf("successfully add a shot: '%s'", shot))
}
