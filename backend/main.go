package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Host 结构体用于存储主机信息
type Host struct {
	ID          int       `json:"id"`
	Hostname    string    `json:"hostname"`
	IP          string    `json:"ip"`
	Group       string    `json:"group"`
	Status      string    `json:"status"`
	LastCheck   time.Time `json:"last_check"`
	Description string    `json:"description"`
}

type AnsibleRequest struct {
	Playbook  string                 `json:"playbook"`
	Inventory string                 `json:"inventory"`
	Variables map[string]interface{} `json:"variables"`
}

type AnsibleResponse struct {
	Output string `json:"output"`
}

// 添加新的结构体
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusComplete  TaskStatus = "complete"
	TaskStatusFailed    TaskStatus = "failed"
)

type TaskLog struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	Message   string    `json:"message"`
	Level     string    `json:"level"` // info, warning, error
	Timestamp time.Time `json:"timestamp"`
}

// 更新 Task 结构体
type Task struct {
	ID        int         `json:"id"`
	Playbook  string      `json:"playbook"`
	Inventory string      `json:"inventory"`
	Output    string      `json:"output"`
	Status    TaskStatus  `json:"status"`
	Progress  int         `json:"progress"` // 0-100
	StartTime time.Time   `json:"start_time"`
	EndTime   *time.Time  `json:"end_time,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

const (
	TEMPLATES_DIR = "./templates"  // 模板文件存储目录
	PLAYBOOK_DIR  = "/playbooks"   // playbook模板子目录
	INVENTORY_DIR = "/inventories" // inventory模板子目录
)

type PlaybookTemplate struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	Type        string    `json:"type"`
	Variables   []string  `json:"variables"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Filename    string    `json:"filename"`  // 添加文件名字段
}

// 添加新的结构体
type PlaybookCheckRequest struct {
	Playbook  string                 `json:"playbook"`
	Inventory string                 `json:"inventory"`
	Variables map[string]interface{} `json:"variables"`
}

type PlaybookCheckResponse struct {
	Valid         bool   `json:"valid"`
	AffectedHosts []Host `json:"affected_hosts"`
	Message       string `json:"message"`
}

// 添加新的结构体
type Role struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tasks       []Task    `json:"tasks"`
	Variables   []string  `json:"variables"`
	Dependencies []string `json:"dependencies"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 添加新的结构体
type File struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // inventory, playbook, config, etc.
	Content     string    `json:"content"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 添加新的结构体
type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
	NotificationTypeSuccess NotificationType = "success"
)

type Notification struct {
	ID        int              `json:"id"`
	Type      NotificationType `json:"type"`
	Message   string          `json:"message"`
	Read      bool            `json:"read"`
	CreatedAt time.Time       `json:"created_at"`
}

var (
	tasks      []Task
	taskID     int
	tasksMutex sync.Mutex
	hosts      []Host
	hostID     int
	hostsMutex sync.Mutex
	templates      []PlaybookTemplate
	templateID     int
	templatesMutex sync.Mutex
	taskLogs    []TaskLog
	taskLogID   int
	logsMutex   sync.Mutex
	roles      []Role
	roleID     int
	rolesMutex sync.Mutex
	files      []File
	fileID     int
	filesMutex sync.Mutex
	notifications      []Notification
	notificationID     int
	notificationsMutex sync.Mutex
)

// 添加新的处理函数
func addTaskLog(taskID int, message string, level string) {
	logsMutex.Lock()
	defer logsMutex.Unlock()

	taskLogID++
	log := TaskLog{
		ID:        taskLogID,
		TaskID:    taskID,
		Message:   message,
		Level:     level,
		Timestamp: time.Now(),
	}
	taskLogs = append(taskLogs, log)
}

func getTaskLogsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "Missing task_id parameter", http.StatusBadRequest)
		return
	}

	logsMutex.Lock()
	defer logsMutex.Unlock()

	var filteredLogs []TaskLog
	for _, log := range taskLogs {
		if fmt.Sprintf("%d", log.TaskID) == taskID {
			filteredLogs = append(filteredLogs, log)
		}
	}

	json.NewEncoder(w).Encode(filteredLogs)
}

// 更新 runAnsibleHandler
func runAnsibleHandler(w http.ResponseWriter, r *http.Request) {
	// Golang 日志输出到终端
	fmt.Printf("[Go] 开始处理 Ansible 请求\n")

	// 添加 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		fmt.Printf("[Go] 处理 OPTIONS 请求\n")
		return
	}

	var req AnsibleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Playbook == "" || req.Inventory == "" {
		fmt.Printf("[Go] 请求参数无效: %v\n", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 创建临时目录
	tmpDir, err := ioutil.TempDir("", "ansible-*")
	if err != nil {
		fmt.Printf("[Go] 创建临时目录失败: %v\n", err)
		http.Error(w, "Failed to create temp directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)

	// 保存 playbook 到临时文件
	playbookFile := filepath.Join(tmpDir, "playbook.yml")
	if err := ioutil.WriteFile(playbookFile, []byte(req.Playbook), 0644); err != nil {
		fmt.Printf("[Go] 保存 playbook 文件失败: %v\n", err)
		http.Error(w, "Failed to save playbook file", http.StatusInternalServerError)
		return
	}

	// 保存 inventory 到临时文件
	inventoryFile := filepath.Join(tmpDir, "inventory.ini")
	if err := ioutil.WriteFile(inventoryFile, []byte(req.Inventory), 0644); err != nil {
		fmt.Printf("[Go] 保存 inventory 文件失败: %v\n", err)
		http.Error(w, "Failed to save inventory file", http.StatusInternalServerError)
		return
	}

	fmt.Printf("[Go] 临时文件已创建:\nPlaybook: %s\nInventory: %s\n", playbookFile, inventoryFile)

	// 打印文件内容用于调试
	fmt.Printf("[Go] Playbook 内容:\n%s\n", req.Playbook)
	fmt.Printf("[Go] Inventory 内容:\n%s\n", req.Inventory)

	// 设置响应头以支持 Server-Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Printf("[Go] 当前连接不支持流式传输\n")
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	tasksMutex.Lock()
	taskID++
	task := Task{
		ID:        taskID,
		Playbook:  req.Playbook,
		Inventory: req.Inventory,
		Status:    TaskStatusPending,
		Progress:  0,
		StartTime: time.Now(),
		Timestamp: time.Now(),
	}
	tasks = append(tasks, task)
	tasksMutex.Unlock()

	fmt.Printf("[Go] 创建新任务 #%d\n", task.ID)

	// 执行 ansible-playbook 命令
	cmd := exec.Command("ansible-playbook", "-i", inventoryFile, playbookFile)

	// 设置工作目录为临时目录
	cmd.Dir = tmpDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("[Go] 创建标准输出管道失败: %v\n", err)
		fmt.Fprintf(w, "data: ERROR: Failed to create stdout pipe: %v\n\n", err)
		flusher.Flush()
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("[Go] 创建标准错误管道失败: %v\n", err)
		fmt.Fprintf(w, "data: ERROR: Failed to create stderr pipe: %v\n\n", err)
		flusher.Flush()
		return
	}

	// 开始执行命令
	if err := cmd.Start(); err != nil {
		fmt.Printf("[Go] 启动命令失败: %v\n", err)
		fmt.Fprintf(w, "data: ERROR: Failed to start command: %v\n\n", err)
		flusher.Flush()
		return
	}

	// 更新任务状态为运行中
	tasksMutex.Lock()
	for i := range tasks {
		if tasks[i].ID == task.ID {
			tasks[i].Status = TaskStatusRunning
			break
		}
	}
	tasksMutex.Unlock()

	fmt.Printf("[Go] 任务 #%d 开始执行\n", task.ID)

	// 读取标准输出
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			// Ansible 输出发送到客户端
			fmt.Fprintf(w, "data: %s\n\n", line)
			flusher.Flush()
			// Golang 日志输出到终端
			fmt.Printf("[Ansible] %s\n", line)
		}
	}()

	// 读取标准错误
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			// Ansible 错误输出发送到客户端
			fmt.Fprintf(w, "data: ERROR: %s\n\n", line)
			flusher.Flush()
			// Golang 日志输出到终端
			fmt.Printf("[Ansible Error] %s\n", line)
		}
	}()

	// 等待命令完成
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("[Go] 任务 #%d 执行失败: %v\n", task.ID, err)
		fmt.Fprintf(w, "data: ERROR: Command failed: %v\n\n", err)
		tasksMutex.Lock()
		for i := range tasks {
			if tasks[i].ID == task.ID {
				tasks[i].Status = TaskStatusFailed
				tasks[i].Output = err.Error()
				endTime := time.Now()
				tasks[i].EndTime = &endTime
				break
			}
		}
		tasksMutex.Unlock()
		addNotification(NotificationTypeError, fmt.Sprintf("任务 #%d 执行失败", task.ID))
	} else {
		fmt.Printf("[Go] 任务 #%d 执行成功\n", task.ID)
		fmt.Fprintf(w, "data: Command completed successfully\n\n")
		tasksMutex.Lock()
		for i := range tasks {
			if tasks[i].ID == task.ID {
				tasks[i].Status = TaskStatusComplete
				tasks[i].Progress = 100
				endTime := time.Now()
				tasks[i].EndTime = &endTime
				break
			}
		}
		tasksMutex.Unlock()
		addNotification(NotificationTypeSuccess, fmt.Sprintf("任务 #%d 执行成功", task.ID))
	}
	flusher.Flush()
}

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	tasksMutex.Lock()
	defer tasksMutex.Unlock()
	json.NewEncoder(w).Encode(tasks)
}

// 添加新的处理函数
func addHostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	var host Host
	if err := json.NewDecoder(r.Body).Decode(&host); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hostsMutex.Lock()
	hostID++
	host.ID = hostID
	host.Status = "unknown"
	host.LastCheck = time.Now()
	hosts = append(hosts, host)
	hostsMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(host)
}

func getHostsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	hostsMutex.Lock()
	defer hostsMutex.Unlock()
	json.NewEncoder(w).Encode(hosts)
}

func checkHostHealth(host *Host) {
	cmd := exec.Command("ansible", host.Hostname, "-m", "ping")
	err := cmd.Run()
	
	host.LastCheck = time.Now()
	if err != nil {
		host.Status = "unhealthy"
	} else {
		host.Status = "healthy"
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	hostsMutex.Lock()
	defer hostsMutex.Unlock()

	for i := range hosts {
		checkHostHealth(&hosts[i])
	}

	json.NewEncoder(w).Encode(hosts)
}

// 添加新的处理函数
func addTemplateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	var template PlaybookTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 根据类型确定文件扩展名和目录
	var ext, dir string
	if template.Type == "playbook" {
		ext = ".yml"
		dir = PLAYBOOK_DIR
	} else if template.Type == "inventory" {
		ext = ".ini"
		dir = INVENTORY_DIR
	} else {
		http.Error(w, "Invalid template type", http.StatusBadRequest)
		return
	}

	// 生成文件名
	filename := fmt.Sprintf("%s%s", template.Name, ext)
	filepath := filepath.Join(TEMPLATES_DIR, dir, filename)

	// 保存文件
	if err := ioutil.WriteFile(filepath, []byte(template.Content), 0644); err != nil {
		http.Error(w, "Failed to save template file", http.StatusInternalServerError)
		return
	}

	templatesMutex.Lock()
	templateID++
	template.ID = templateID
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	template.Filename = filename
	templates = append(templates, template)
	templatesMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

func getTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		return
	}

	templateType := r.URL.Query().Get("type")
	
	templatesMutex.Lock()
	defer templatesMutex.Unlock()
	
	var filteredTemplates []PlaybookTemplate
	for _, template := range templates {
		if templateType == "" || template.Type == templateType {
			filteredTemplates = append(filteredTemplates, template)
		}
	}
	
	json.NewEncoder(w).Encode(filteredTemplates)
}

// 添加新的处理函数
func checkPlaybookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	var req PlaybookCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 使用 ansible-playbook --check 模式来验证 playbook
	cmd := exec.Command("ansible-playbook", "--check", "-i", req.Inventory, req.Playbook)
	output, err := cmd.CombinedOutput()

	response := PlaybookCheckResponse{
		Valid:         err == nil,
		AffectedHosts: []Host{},
		Message:       string(output),
	}

	// 如果验证通过，获取受影响的主机
	if err == nil {
		hostsMutex.Lock()
		for _, host := range hosts {
			// 这里需要根据实际情况来判断哪些主机会受到影响
			// 可以通过解析 inventory 文件或其他方式来实现
			response.AffectedHosts = append(response.AffectedHosts, host)
		}
		hostsMutex.Unlock()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 添加新的处理函数
func addRoleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	var role Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	rolesMutex.Lock()
	roleID++
	role.ID = roleID
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	roles = append(roles, role)
	rolesMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(role)
}

func getRolesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	rolesMutex.Lock()
	defer rolesMutex.Unlock()
	json.NewEncoder(w).Encode(roles)
}

// 添加新的处理函数
func addFileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	var file File
	if err := json.NewDecoder(r.Body).Decode(&file); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	filesMutex.Lock()
	fileID++
	file.ID = fileID
	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()
	files = append(files, file)
	filesMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(file)
}

func getFilesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	fileType := r.URL.Query().Get("type")

	filesMutex.Lock()
	defer filesMutex.Unlock()

	if fileType == "" {
		json.NewEncoder(w).Encode(files)
		return
	}

	var filteredFiles []File
	for _, file := range files {
		if file.Type == fileType {
			filteredFiles = append(filteredFiles, file)
		}
	}
	json.NewEncoder(w).Encode(filteredFiles)
}

func updateFileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	var updatedFile File
	if err := json.NewDecoder(r.Body).Decode(&updatedFile); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	filesMutex.Lock()
	defer filesMutex.Unlock()

	for i := range files {
		if files[i].ID == updatedFile.ID {
			updatedFile.CreatedAt = files[i].CreatedAt
			updatedFile.UpdatedAt = time.Now()
			files[i] = updatedFile
			json.NewEncoder(w).Encode(updatedFile)
			return
		}
	}

	http.Error(w, "File not found", http.StatusNotFound)
}

// 添加新的处理函数
func addNotification(notificationType NotificationType, message string) {
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()

	notificationID++
	notification := Notification{
		ID:        notificationID,
		Type:      notificationType,
		Message:   message,
		Read:      false,
		CreatedAt: time.Now(),
	}
	notifications = append(notifications, notification)
}

func getNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()
	json.NewEncoder(w).Encode(notifications)
}

func markNotificationReadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()

	for i := range notifications {
		if notifications[i].ID == req.ID {
			notifications[i].Read = true
			json.NewEncoder(w).Encode(notifications[i])
			return
		}
	}

	http.Error(w, "Notification not found", http.StatusNotFound)
}

// 初始化模板目录
func initTemplatesDirs() error {
	dirs := []string{
		filepath.Join(TEMPLATES_DIR, PLAYBOOK_DIR),
		filepath.Join(TEMPLATES_DIR, INVENTORY_DIR),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// 从文件系统加载模板
func loadTemplatesFromFiles() error {
	templates = []PlaybookTemplate{} // 清空内存中的模板
	
	// 加载 playbook 模板
	playbookFiles, err := ioutil.ReadDir(filepath.Join(TEMPLATES_DIR, PLAYBOOK_DIR))
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Join(TEMPLATES_DIR, PLAYBOOK_DIR), 0755); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	
	for _, file := range playbookFiles {
		if filepath.Ext(file.Name()) == ".yml" || filepath.Ext(file.Name()) == ".yaml" {
			content, err := ioutil.ReadFile(filepath.Join(TEMPLATES_DIR, PLAYBOOK_DIR, file.Name()))
			if err != nil {
				continue
			}
			
			templateID++
			template := PlaybookTemplate{
				ID:        templateID,
				Name:      strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())),
				Content:   string(content),
				Type:      "playbook",
				Filename:  file.Name(),
				CreatedAt: file.ModTime(),
				UpdatedAt: file.ModTime(),
			}
			templates = append(templates, template)
		}
	}
	
	// 加载 inventory 模板
	inventoryFiles, err := ioutil.ReadDir(filepath.Join(TEMPLATES_DIR, INVENTORY_DIR))
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Join(TEMPLATES_DIR, INVENTORY_DIR), 0755); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	
	for _, file := range inventoryFiles {
		if filepath.Ext(file.Name()) == ".ini" {
			content, err := ioutil.ReadFile(filepath.Join(TEMPLATES_DIR, INVENTORY_DIR, file.Name()))
			if err != nil {
				continue
			}
			
			templateID++
			template := PlaybookTemplate{
				ID:        templateID,
				Name:      strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())),
				Content:   string(content),
				Type:      "inventory",
				Filename:  file.Name(),
				CreatedAt: file.ModTime(),
				UpdatedAt: file.ModTime(),
			}
			templates = append(templates, template)
		}
	}
	
	return nil
}

func updateTemplateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	fmt.Printf("[Go] 开始处理模板更新请求\n")

	var template PlaybookTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		fmt.Printf("[Go] 解析请求体失败: %v\n", err)
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	fmt.Printf("[Go] 收到更新请求: ID=%d, Name=%s, Type=%s\n", template.ID, template.Name, template.Type)

	// 根据类型确定目录
	var dir string
	if template.Type == "playbook" {
		dir = PLAYBOOK_DIR
	} else if template.Type == "inventory" {
		dir = INVENTORY_DIR
	} else {
		fmt.Printf("[Go] 无效的模板类型: %s\n", template.Type)
		http.Error(w, "Invalid template type", http.StatusBadRequest)
		return
	}

	// 更新文件
	filepath := filepath.Join(TEMPLATES_DIR, dir, template.Filename)
	fmt.Printf("[Go] 准备更新文件: %s\n", filepath)

	if err := ioutil.WriteFile(filepath, []byte(template.Content), 0644); err != nil {
		fmt.Printf("[Go] 保存文件失败: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to save template file: %v", err), http.StatusInternalServerError)
		return
	}

	templatesMutex.Lock()
	found := false
	for i := range templates {
		if templates[i].ID == template.ID {
			fmt.Printf("[Go] 找到要更新的模板: ID=%d\n", template.ID)
			template.UpdatedAt = time.Now()
			templates[i] = template
			json.NewEncoder(w).Encode(template)
			found = true
			break
		}
	}
	templatesMutex.Unlock()

	if !found {
		fmt.Printf("[Go] 未找到要更新的模板: ID=%d\n", template.ID)
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	fmt.Printf("[Go] 模板更新成功: ID=%d\n", template.ID)
}

func main() {
	// 初始化模板目录
	if err := initTemplatesDirs(); err != nil {
		fmt.Printf("Failed to initialize template directories: %v\n", err)
		return
	}

	// 加载已有模板
	if err := loadTemplatesFromFiles(); err != nil {
		fmt.Printf("Failed to load templates from files: %v\n", err)
		return
	}

	// 打印已加载的模板信息
	fmt.Printf("Loaded %d templates\n", len(templates))
	for _, t := range templates {
		fmt.Printf("Template: %s (Type: %s)\n", t.Name, t.Type)
	}

	http.HandleFunc("/run", runAnsibleHandler)
	http.HandleFunc("/tasks", getTasksHandler)
	http.HandleFunc("/hosts", getHostsHandler)
	http.HandleFunc("/hosts/add", addHostHandler)
	http.HandleFunc("/hosts/health", healthCheckHandler)
	http.HandleFunc("/templates", getTemplatesHandler)
	http.HandleFunc("/templates/add", addTemplateHandler)
	http.HandleFunc("/templates/update", updateTemplateHandler)
	http.HandleFunc("/tasks/logs", getTaskLogsHandler)
	http.HandleFunc("/playbook/check", checkPlaybookHandler)
	http.HandleFunc("/roles", getRolesHandler)
	http.HandleFunc("/roles/add", addRoleHandler)
	http.HandleFunc("/files", getFilesHandler)
	http.HandleFunc("/files/add", addFileHandler)
	http.HandleFunc("/files/update", updateFileHandler)
	http.HandleFunc("/notifications", getNotificationsHandler)
	http.HandleFunc("/notifications/read", markNotificationReadHandler)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}