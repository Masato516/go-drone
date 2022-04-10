package models

import (
	"sync"
	"time"
)

type BaseCourse interface {
	Start()
	Stop()
	Run()
	UpdateElapsed()
}

type Course struct {
	Name      string        `json:name`
	Status    int           `json:status`
	IsRunning bool          `json:is_running`
	StartTime time.Time     `json:start_time`
	Elasped   time.Duration `json:elasped`
	Drone     *DroneManager `json:"-"`
	mux       sync.Mutex    `json:"-"`
}

func (c *Course) Start() {
	if c.IsRunning {
		return
	}
	c.IsRunning = true
	c.StartTime = time.Now()
}

func (c *Course) Stop() {
	if !c.IsRunning {
		return
	}
	c.IsRunning = false
	c.Status = 0
}

func (c *Course) UpdateElapsed() {
	if !c.IsRunning {
		return
	}
	c.Elasped = time.Since(c.StartTime)
}

type CourseA struct {
	Course
}

func (c *CourseA) Run() {
	c.mux.Lock()
	defer c.mux.Unlock()
	if !c.IsRunning {
		return
	}

	c.Status++
	switch c.Status {
	case 1:
		c.Drone.TakeOff()
	case 10:
		c.Drone.Clockwise(30)
	case 15:
		c.Drone.CounterClockwise(30)
	case 20:
		c.Drone.Clockwise(30)
	case 25:
		c.Drone.CounterClockwise(30)
	case 30:
		c.Drone.Hover()
	case 35:
		c.Drone.FrontFlip()
	case 45:
		c.Drone.BackFlip()
	case 55:
		c.Drone.Land()
		c.Stop()
	}
	c.UpdateElapsed()
}

type CourseB struct {
	Course
}

func (c *CourseB) Run() {
	c.mux.Lock()
	defer c.mux.Unlock()

	if !c.IsRunning {
		return
	}

	c.Status++
	switch c.Status {
	case 1:
		c.Drone.TakeOff()
	case 10:
		c.Drone.FrontFlip()
	case 20:
		c.Drone.FrontFlip()
		if time.Since(c.StartTime) < 10 {
			c.Status = 35
		}
	case 30:
		c.Drone.Clockwise(30)
	case 40:
		c.Drone.Hover()
	case 50:
		c.Drone.Land()
		c.Stop()
	}
	c.UpdateElapsed()
}

func NewDefaultCourse(droneManager *DroneManager) map[int]BaseCourse {
	var a, b BaseCourse
	a = &CourseA{Course{Name: "Course A", Drone: droneManager}}
	b = &CourseB{Course{Name: "Course B", Drone: droneManager}}
	return map[int]BaseCourse{1: a, 2: b}
}
