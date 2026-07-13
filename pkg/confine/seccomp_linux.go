//go:build linux

package confine

import seccomp "github.com/seccomp/libseccomp-golang"
import "golang.org/x/sys/unix"

// applyNetwork 用 seccomp 禁止外联:在 socket() 上做参数级过滤,
// 当地址族为 AF_INET / AF_INET6 时返回 EPERM,而放行 AF_UNIX,
// 从而切断 IPv4/IPv6 网络但不破坏本地 Unix 域套接字通信。
//
// requireEnforcement 为 false 时,过滤器加载失败(平台/内核不支持)将被容忍。
func applyNetwork(requireEnforcement bool) error {
	filter, err := seccomp.NewFilter(seccomp.ActAllow) // 默认放行,仅对出网 socket 设例外
	if err != nil {
		return err
	}
	defer filter.Release()

	if err := filter.SetNoNewPrivsBit(true); err != nil {
		return err
	}
	// 注意:本绑定不暴露 TSYNC 开关,也无需手动设置。
	// Load() 内部会自动启用 SCMP_FLTATR_CTL_TSYNC,把过滤器同步到
	// Go 运行时的所有 OS 线程(内核 >= 3.17 且 libseccomp >= 2.2 时生效),
	// 防止其它线程绕过限制。同步失败时 Load() 会返回错误。

	socketID, err := seccomp.GetSyscallFromName("socket")
	if err != nil {
		return err
	}
	denied := seccomp.ActErrno.SetReturnCode(int16(unix.EPERM))

	for _, family := range []uint64{unix.AF_INET, unix.AF_INET6} {
		cond, err := seccomp.MakeCondition(0, seccomp.CompareEqual, family) // arg0 == family
		if err != nil {
			return err
		}
		if err := filter.AddRuleConditional(socketID, denied, []seccomp.ScmpCondition{cond}); err != nil {
			return err
		}
	}

	if err := filter.Load(); err != nil {
		if requireEnforcement {
			return err
		}
		return nil // best-effort:不支持则放行
	}
	return nil
}
