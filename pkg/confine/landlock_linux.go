//go:build linux

package confine

import landlock "github.com/landlock-lsm/go-landlock/landlock"

// applyFilesystem 用 Landlock 施加文件系统约束:全盘(或指定范围)只读,
// 仅 WritableDirs 可写。go-landlock 负责 ABI 协商与 NO_NEW_PRIVS。
func applyFilesystem(p Policy) error {
	var rules []landlock.Rule

	if len(p.ReadableDirs) == 0 {
		rules = append(rules, landlock.RODirs("/"))
	} else {
		// 列出的只读目录可能部分不存在,忽略缺失项以增强健壮性。
		rules = append(rules, landlock.RODirs(p.ReadableDirs...).IgnoreIfMissing())
	}

	for _, dir := range p.WritableDirs {
		rules = append(rules, landlock.RWDirs(dir))
	}

	// 许多程序需要写 /dev/null。
	rules = append(rules, landlock.RWFiles("/dev/null").IgnoreIfMissing())

	cfg := landlock.V5 // 请求最新 ABI;库会向下兼容到内核实际支持的版本
	if !p.RequireEnforcement {
		cfg = cfg.BestEffort() // 内核过老时降级而非报错
	}
	return cfg.RestrictPaths(rules...)
}
