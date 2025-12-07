package storage

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

const dataDir = "data"

// TaskData хранит данные задачи: id, список ссылок и результаты проверки.
type TaskData struct {
	ID     int               `json:"id"`
	Links  []string          `json:"links"`
	Result map[string]string `json:"result"`
}

// Memory — простое файловое хранилище в памяти с сохранением в JSON-файлы.
// Использует mutex для безопасного доступа из нескольких горутин.
type Memory struct {
	mu     sync.RWMutex
	lastID int
	tasks  map[int]*TaskData
}

// New создает новое хранилище, создаёт директорию `data` и загружает
// ранее сохранённые данные, если они присутствуют.
func New() *Memory {
	_ = os.MkdirAll(dataDir, 0755)

	m := &Memory{
		lastID: 1,
		tasks:  make(map[int]*TaskData),
	}

	m.load()
	return m
}

// CreateTask создаёт новую запись задачи и сразу сохраняет состояние.
func (m *Memory) CreateTask(links []string) *TaskData {
	m.mu.Lock()
	defer m.mu.Unlock()

	t := &TaskData{
		ID:     m.lastID,
		Links:  links,
		Result: make(map[string]string),
	}

	// инициализируем результаты с начальным статусом "checking"
	for _, link := range links {
		t.Result[link] = "checking"
	}

	m.tasks[m.lastID] = t
	m.lastID++

	m.save()
	return t
}

// SaveResult сохраняет статус конкретной ссылки в задаче и пишет данные на диск.
func (m *Memory) SaveResult(id int, link, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if t, ok := m.tasks[id]; ok {
		t.Result[link] = status
	}

	m.save()
}

// GetMany возвращает результаты по множеству идентификаторов задач.
func (m *Memory) GetMany(ids []int) map[int]map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make(map[int]map[string]string)
	for _, id := range ids {
		if t, ok := m.tasks[id]; ok {
			res[id] = t.Result
		}
	}
	return res
}

// save записывает текущее состояние в `data/tasks.json` и `data/next_id.txt`.
func (m *Memory) save() {
	data, err := json.MarshalIndent(m.tasks, "", "  ")
	if err != nil {
		log.Printf("failed to marshal tasks: %v", err)
		return
	}
	if err := os.WriteFile(filepath.Join(dataDir, "tasks.json"), data, 0644); err != nil {
		log.Printf("failed to write tasks.json: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dataDir, "next_id.txt"), []byte(strconv.Itoa(m.lastID)), 0644); err != nil {
		log.Printf("failed to write next_id.txt: %v", err)
	}
}

// load пытается прочитать предыдущие файлы состояния, если они существуют.
func (m *Memory) load() {
	data, err := os.ReadFile(filepath.Join(dataDir, "tasks.json"))
	if err == nil {
		if err := json.Unmarshal(data, &m.tasks); err != nil {
			log.Printf("failed to unmarshal tasks.json: %v", err)
			m.tasks = make(map[int]*TaskData)
		}
	}

	idData, err := os.ReadFile(filepath.Join(dataDir, "next_id.txt"))
	if err == nil {
		if n, err := strconv.Atoi(string(idData)); err == nil {
			m.lastID = n
		}
	}
}
