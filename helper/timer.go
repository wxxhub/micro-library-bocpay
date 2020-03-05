package helper

import (
	"time"
)

type Timer struct {
	duration map[string]*duration
}

func (t *Timer) Start(name string) {
	if t.duration == nil {
		t.duration = make(map[string]*duration)
	}
	t.duration[name] = new(duration)
	t.duration[name].Start()
}

func (t *Timer) End(name string) {
	if t.duration == nil {
		t.duration = make(map[string]*duration)
	}
	if _, ok := t.duration[name]; ok {
		t.duration[name].End()
	}
}

func (t *Timer) Duration(name string) time.Duration {
	if t.duration == nil {
		t.duration = make(map[string]*duration)
	}
	return t.duration[name].Duration()
}

func (t *Timer) Calculation() (map[string]string) {
	//转换时间格式
	var strTimer = make(map[string]string)
	var recordProfiler time.Duration
	allTimer := t.Duration("allTimer")
	for key, val := range t.duration {
		dura := val.Duration()
		//精确到毫秒值
		strTimer[key] = dura.Round(time.Millisecond).String()
		if key != "allTimer" {
			//Timer标记过的总时间
			recordProfiler += dura
		}
	}
	//其余没有记录的百分比耗时
	strTimer["other"] = (allTimer - recordProfiler).Round(time.Millisecond).String()
	return strTimer
}

type duration struct {
	start    time.Time
	end      time.Time
	duration time.Duration
}

func (d *duration) Start() {
	d.start = time.Now()
}

func (d *duration) End() {
	d.end = time.Now()
}

func (d *duration) Duration() time.Duration {
	if d.end.IsZero() {
		d.End()
	}
	return d.end.Sub(d.start)
}
