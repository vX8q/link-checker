package service

import (
	"link/internal/storage"
	"time"
)

// Task представляет собой задачу проверки набора ссылок.
// Поля отражают идентификатор задачи, список ссылок, результат
// проверки (map: ссылка -> статус) и флаг завершения.
type Task struct {
	ID     int               `json:"id"`
	Links  []string          `json:"links"`
	Result map[string]string `json:"result"`
	Done   bool              `json:"done"`
}

// TaskService управляет созданием задач проверки и пулом воркеров,
// которые выполняют проверку ссылок в фоне.
type TaskService struct {
	st   *storage.Memory
	work chan *Task
	stop chan struct{}
}

// NewTaskService создаёт новый сервис задач с указанным хранилищем.
// Запускает фиксированное количество фоновых воркеров для обработки задач.
func NewTaskService(st *storage.Memory) *TaskService {
	s := &TaskService{
		st:   st,
		work: make(chan *Task, 100),
		stop: make(chan struct{}),
	}

	// запустим пул воркеров
	for i := 0; i < 5; i++ {
		go s.worker()
	}

	return s
}

// Create создаёт новую задачу проверки ссылок и ставит её в очередь обработки.
// Возвращает ID задачи и начальные статусы (map ссылка->статус).
func (s *TaskService) Create(links []string) (int, map[string]string) {
	taskData := s.st.CreateTask(links)
	t := &Task{
		ID:     taskData.ID,
		Links:  taskData.Links,
		Result: taskData.Result,
		Done:   false,
	}

	// помещаем задачу в очередь для фоновой обработки
	s.work <- t

	return t.ID, taskData.Result
}

// worker — фоновый воркер, который читает задачи из канала `work`
// и вызывает метод проверки ссылок. Завершается при закрытии `stop`.
func (s *TaskService) worker() {
	for {
		select {
		case t := <-s.work:
			s.checkLinks(t)
		case <-s.stop:
			return
		}
	}
}

// Shutdown корректно останавливает сервис, послав сигнал завершения воркерам.
func (s *TaskService) Shutdown() {
	close(s.stop)
}

// GetForReport возвращает результаты для списка ID задач — используется
// для формирования отчетов по проверкам.
func (s *TaskService) GetForReport(ids []int) map[int]map[string]string {
	return s.st.GetMany(ids)
}

// WaitAndGetResults ждёт завершения проверки ссылок для задачи и возвращает результаты.
// Использует timeout для ограничения времени ожидания.
func (s *TaskService) WaitAndGetResults(taskID int, maxWait time.Duration) map[string]string {
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		results := s.st.GetMany([]int{taskID})
		if taskResults, ok := results[taskID]; ok {
			// проверим, все ли результаты обновлены (нет "checking")
			allDone := true
			for _, status := range taskResults {
				if status == "checking" {
					allDone = false
					break
				}
			}
			if allDone && len(taskResults) > 0 {
				return taskResults
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	// если timeout истёк, вернём то, что есть
	results := s.st.GetMany([]int{taskID})
	if taskResults, ok := results[taskID]; ok {
		return taskResults
	}
	return make(map[string]string)
}
