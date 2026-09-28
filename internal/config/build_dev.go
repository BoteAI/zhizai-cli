//go:build dev

package config

// devBuild 仅在 go build -tags dev 的开发二进制中为 true，
// 此时 ZHIZAI_DEV=1 + ZHIZAI_ENV / api_url 等切换项才会生效。
const devBuild = true
