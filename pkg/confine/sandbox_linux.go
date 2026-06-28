//go:build linux

// Package confine 把当前进程约束在内核级沙箱内。
//
// 它组合两层 Linux 内核机制:
//   - Landlock:按文件路径限制访问(全盘只读,仅指定目录可写)。
//   - seccomp:按系统调用及其参数限制(默认禁止外联 socket)。
//
// 两者一旦施加都不可撤销,且跨 execve 继承。因此 Apply 应在一个
// 一次性的、专门用于执行受限逻辑的进程里调用,而非长期存活的进程。
//
// NO_NEW_PRIVS 由底层库(go-landlock / libseccomp)负责开启,无需显式设置。
package confine

import (
	"fmt"
	"runtime"
)

// Policy 描述施加的约束。零值表示最严:全盘只读、断网、强制 enforcement。
type Policy struct {
	// WritableDirs 列出允许写入的目录;其余路径只读。目录须已存在。
	WritableDirs []string

	// ReadableDirs 限制可读范围。为空表示全盘可读(危害主要来自写与外联,
	// 限制读往往得不偿失且易破坏工具)。列出的目录若不存在会被忽略。
	ReadableDirs []string

	// AllowNetwork 为 true 时不安装网络过滤器。
	AllowNetwork bool

	// RequireEnforcement 为 true 时,内核不支持所需隔离则返回错误(fail-closed);
	// 为 false 时尽力施加、不支持就跳过(fail-open)。
	RequireEnforcement bool
}

// Apply 在当前 OS 线程上施加 p 描述的约束。
//
// 调用后约束不可撤销。Apply 会锁定当前 goroutine 到其 OS 线程,
// 以保证"施加约束"与"随后执行受限逻辑"发生在同一线程上;
// 调用方不应在 Apply 之后解锁该线程。
func Apply(p Policy) error {
	runtime.LockOSThread()

	if err := applyFilesystem(p); err != nil {
		return fmt.Errorf("confine: filesystem: %w", err)
	}
	if !p.AllowNetwork {
		if err := applyNetwork(p.RequireEnforcement); err != nil {
			return fmt.Errorf("confine: network: %w", err)
		}
	}
	return nil
}
