package addon

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"

	"gopkg.in/yaml.v3"
)

// Pack 将已安装的扩展打包为可分发的 ZIP 文件。
// 打包结构与 Install 期望的 ZIP 结构一致：
//
//	{name}.zip
//	  ├── addon.yaml
//	  ├── server/          ← server/addons/{name}/ 下除 addon.yaml 外的所有文件
//	  ├── web/             ← web/src/addons/{name}/ 下的所有文件
//	  ├── install/         ← server/addons/{name}/install/
//	  ├── uninstall/       ← server/addons/{name}/uninstall/
//	  └── upgrade/         ← server/addons/{name}/upgrade/（如果有）
func Pack(ctx context.Context, name string) error {
	projectRoot := getProjectRoot()

	if name == "" {
		installed := scanInstalledAddonsForPack(projectRoot)
		if len(installed) == 0 {
			fmt.Println("  未发现已安装的扩展")
			return nil
		}
		fmt.Println("  已安装的扩展：")
		for i, a := range installed {
			fmt.Printf("    [%d] %s (%s)\n", i+1, a.Name, a.Title)
		}
		fmt.Println()
		choice := gcmd.Scan("  请选择要打包的扩展编号: ")
		choice = strings.TrimSpace(choice)
		idx := 0
		if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > len(installed) {
			return fmt.Errorf("无效的选择")
		}
		name = installed[idx-1].Name
	}

	serverDir := filepath.Join(projectRoot, "server", "addons", name)
	webDir := filepath.Join(projectRoot, "web", "src", "addons", name)
	yamlPath := filepath.Join(serverDir, "addon.yaml")

	if !gfile.Exists(yamlPath) {
		return fmt.Errorf("扩展 %q 不存在或缺少 addon.yaml（路径: %s）", name, yamlPath)
	}

	raw, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("读取 addon.yaml 失败: %v", err)
	}
	var meta AddonMeta
	if err := yaml.Unmarshal(raw, &meta); err != nil {
		return fmt.Errorf("解析 addon.yaml 失败: %v", err)
	}

	zipName := fmt.Sprintf("%s.zip", name)
	if meta.Version != "" {
		zipName = fmt.Sprintf("%s-%s.zip", name, meta.Version)
	}

	outputPath := filepath.Join(projectRoot, "server", "addons", zipName)

	fmt.Printf("  打包扩展: %s v%s\n", meta.Title, meta.Version)
	fmt.Printf("  输出文件: %s\n", outputPath)

	zf, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建 ZIP 文件失败: %v", err)
	}
	defer zf.Close()

	w := zip.NewWriter(zf)
	defer w.Close()

	// SQL 目录需要提升到 ZIP 根级别而不是放在 server/ 下
	sqlDirs := map[string]bool{
		"install":   true,
		"uninstall": true,
		"upgrade":   true,
	}

	totalFiles := 0

	// 1. addon.yaml → ZIP 根
	if err := addFileToZip(w, yamlPath, "addon.yaml"); err != nil {
		return fmt.Errorf("写入 addon.yaml 失败: %v", err)
	}
	totalFiles++

	// 2. server/ 目录 → ZIP server/
	if gfile.Exists(serverDir) {
		err := filepath.Walk(serverDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			relPath, _ := filepath.Rel(serverDir, path)
			relPath = filepath.ToSlash(relPath)

			if relPath == "addon.yaml" {
				return nil
			}

			topDir := strings.SplitN(relPath, "/", 2)[0]
			if sqlDirs[topDir] {
				// install/, uninstall/, upgrade/ 提升到 ZIP 根
				if err := addFileToZip(w, path, relPath); err != nil {
					return err
				}
			} else {
				if err := addFileToZip(w, path, "server/"+relPath); err != nil {
					return err
				}
			}
			totalFiles++
			return nil
		})
		if err != nil {
			return fmt.Errorf("打包后端文件失败: %v", err)
		}
	}

	// 3. web/ 目录 → ZIP web/
	if gfile.Exists(webDir) {
		err := filepath.Walk(webDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			relPath, _ := filepath.Rel(webDir, path)
			relPath = filepath.ToSlash(relPath)

			if err := addFileToZip(w, path, "web/"+relPath); err != nil {
				return err
			}
			totalFiles++
			return nil
		})
		if err != nil {
			return fmt.Errorf("打包前端文件失败: %v", err)
		}
	}

	fmt.Printf("  完成: %s (%d 个文件)\n", zipName, totalFiles)
	return nil
}

func addFileToZip(w *zip.Writer, srcPath, zipPath string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = zipPath
	header.Method = zip.Deflate

	writer, err := w.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, f)
	return err
}

type addonBrief struct {
	Name  string
	Title string
}

func scanInstalledAddonsForPack(projectRoot string) []addonBrief {
	addonsDir := filepath.Join(projectRoot, "server", "addons")
	if !gfile.Exists(addonsDir) {
		return nil
	}
	entries, err := os.ReadDir(addonsDir)
	if err != nil {
		return nil
	}
	var result []addonBrief
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		yamlPath := filepath.Join(addonsDir, e.Name(), "addon.yaml")
		if !gfile.Exists(yamlPath) {
			continue
		}
		raw, err := os.ReadFile(yamlPath)
		if err != nil {
			continue
		}
		var meta AddonMeta
		if err := yaml.Unmarshal(raw, &meta); err != nil {
			continue
		}
		title := meta.Title
		if title == "" {
			title = meta.Name
		}
		result = append(result, addonBrief{Name: e.Name(), Title: title})
	}
	return result
}
