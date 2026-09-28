//go:build !dev

package config

// devBuild 在默认构建（含 GitHub Release / npm 发布包）中固定为 false：
// 环境切换能力在编译期被移除，即使设置 ZHIZAI_DEV=1 / ZHIZAI_ENV=test
// 也物理上无法离开生产环境。
const devBuild = false
