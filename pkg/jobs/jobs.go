package jobs

import(
	"sync"
)

type Job struct {
	Do func()
	When uint
	collectedFrames uint
}

func (j *Job) Tick(wg *sync.WaitGroup) {
	defer wg.Done()

	j.collectedFrames++
	if j.collectedFrames == j.When {
		j.collectedFrames = 0
		j.Do()
	}
}

var jobs []Job

func Add(job Job) {
	jobs = append(jobs, job)
}

func TickAllOneFrame() {
	wg := &sync.WaitGroup{}

	wg.Add(len(jobs))
	for i := range jobs {
		go jobs[i].Tick(wg)
	}

	wg.Wait()
}
