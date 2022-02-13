package config

type Zap struct {
	Level         string `json:"level"`          // 级别
	Format        string `json:"format"`         // 输出
	Director      string `json:"director"`       // 日志文件夹
	ShowLine      bool   `json:"show_line"`      // 显示行
	EncodeLevel   string `json:"encode_level"`   // 编码级
	StacktraceKey string `json:"stacktrace_key"` // 栈名
	LogInConsole  bool   `json:"log_in_console"` // 输出控制台
}
