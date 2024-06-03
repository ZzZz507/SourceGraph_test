package review

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	constant "reviewService/utils"
	"strings"
)

// TestInfo 单个函数的测试信息
type TestInfo struct {
	BuildRes   bool   `json:"build_res"`
	BuildInfo  string `json:"build_info,omitempty"`
	RunRes     bool   `json:"run_res"`
	RunInfo    string `json:"run_info,omitempty"`
	AssertRes  bool   `json:"assert_res"`
	AssertInfo string `json:"assert_info,omitempty"`
}

// CovInfo 行覆盖率信息
type CovInfo struct {
	TotalLineNums int `json:"total_line_nums"`
	CovLineNums   int `json:"cov_line_nums"`
}

// StatMetrics 统计数据
type StatMetrics struct {
	TestFileNum      int `json:"test_file_num"`
	TestFuncNum      int `json:"test_func_num"`
	BuildSuccessNum  int `json:"build_success_num"`
	RunSuccessNum    int `json:"run_success_num"`
	AssertSuccessNum int `json:"assert_success_num"`
}

// FileNode 文件节点
type FileNode struct {
	Name        string       `json:"name"`
	Path        string       `json:"path"`
	IsDir       bool         `json:"is_dir"`
	StatMetrics *StatMetrics `json:"stat,omitempty"`
	TestInfo    *TestInfo    `json:"test_info,omitempty"`
	//CovInfo     CovInfo     `json:"cov_info,omitempty"`
	Children []*FileNode `json:"children,omitempty"`
	SubFiles []*FileNode `json:"sub_files,omitempty"`
}

func RunGoTestsInProject(rootPath string) error {
	// 判断路径是否存在
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return fmt.Errorf("error: Path does not exist: %s", rootPath)
	}
	// 初始化根节点
	rootNode := &FileNode{
		Name:        filepath.Base(rootPath),
		Path:        rootPath,
		IsDir:       true,
		StatMetrics: &StatMetrics{},
		Children:    []*FileNode{},
	}
	// 遍历目录
	err := walkDir(rootPath, rootNode)
	if err != nil {
		return fmt.Errorf("error: %s", err)
	}
	// 结构体rootNode——>json
	jsonData, err := json.MarshalIndent(rootNode, "", "  ")
	if err != nil {
		return fmt.Errorf("error: %s", err)
	}
	// 将json写入文件，0644表示权限
	err = os.WriteFile("result.json", jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error: %s", err)
	}
	fmt.Println("result saved to result.json")

	return nil
}

func walkDir(path string, node *FileNode) error {
	// 读取当前路径下的文件切片
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	// 将文件切片按名称分组
	// to do 如果文件和子目录同名，可能会分到同一组，导致递归出错
	groups := make(map[string][]os.DirEntry)
	for _, file := range files {
		base := getBaseName(file.Name())
		groups[base] = append(groups[base], file)
	}
	// 遍历该路径下的所有分组
	for base, group := range groups {
		// 过滤隐藏目录
		if strings.HasPrefix(base, ".") || base == "" {
			continue
		}
		//fmt.Printf("Base: %s, Group: %v\n", base, group)
		if group[0].IsDir() {
			// 初始化子目录节点
			childNode := &FileNode{
				Name:        group[0].Name(),
				Path:        filepath.Join(path, group[0].Name()),
				IsDir:       true,
				StatMetrics: &StatMetrics{},
				Children:    []*FileNode{},
			}
			err := walkDir(childNode.Path, childNode)
			if err != nil {
				return err
			}
			node.Children = append(node.Children, childNode)
			// 累加该目录节点下的统计数据
			node.StatMetrics.TestFileNum += childNode.StatMetrics.TestFileNum
			node.StatMetrics.TestFuncNum += childNode.StatMetrics.TestFuncNum
			node.StatMetrics.BuildSuccessNum += childNode.StatMetrics.BuildSuccessNum
			node.StatMetrics.RunSuccessNum += childNode.StatMetrics.RunSuccessNum
			node.StatMetrics.AssertSuccessNum += childNode.StatMetrics.AssertSuccessNum
			continue
		}
		// 遍历该分组下的所有文件
		for _, file := range group {
			if regexp.MustCompile(fmt.Sprintf(`^%s\.go$`, base)).MatchString(file.Name()) {
				// 初始化测试文件节点信息
				childNode := &FileNode{
					Name:     file.Name(),
					Path:     filepath.Join(path, file.Name()),
					IsDir:    false,
					SubFiles: []*FileNode{},
				}
				node.Children = append(node.Children, childNode)
				node.StatMetrics.TestFileNum++
			} else if regexp.MustCompile(fmt.Sprintf(`^(%s)_[0-100]_test\.go$`, base)).MatchString(file.Name()) {
				// fmt.Println("Matched:", file.Name())
				lastChild := node.Children[len(node.Children)-1]
				// 初始化测试子文件节点信息
				SubFileNode := &FileNode{
					Name:     file.Name(),
					Path:     filepath.Join(path, file.Name()),
					IsDir:    false,
					TestInfo: &TestInfo{},
				}
				// 遍历每个单测文件
				err, testInfo := runTestFile(file.Name(), SubFileNode.Path)
				SubFileNode.TestInfo = testInfo
				// 若该文件编译，执行通过，获取其覆盖率信息
				if err == nil {
					//_, _ = GetOneTestFileCover(file.Name(), SubFileNode.Path)
					fmt.Println("1111111111")
				}
				lastChild.SubFiles = append(lastChild.SubFiles, SubFileNode)
				node.StatMetrics.TestFuncNum++
			}
		}
		// to do 统计该分组的信息，分析逻辑树
		// 分析flag，累加到node.StatMetrics
	}
	return nil
}

func runTestFile(testFile string, path string) (error, *TestInfo) {
	testFilePath := path
	testOriFilePath := regexp.MustCompile(`_[0-100]_test\.go$`).ReplaceAllString(testFilePath, ".go")
	//to do，根据testFile去检索对应的函数名称
	//eg. baseName_2_test.go <==> FuncName
	testInfo := &TestInfo{
		BuildRes:   false,
		BuildInfo:  "",
		RunRes:     false,
		RunInfo:    "",
		AssertRes:  false,
		AssertInfo: "",
	}
	// 运行测试文件
	cmd := exec.Command("go", "test", "-v", testOriFilePath, testFilePath)
	cmd.Dir = constant.RootPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "unknown command") {
			// 未知的命令
			fmt.Println("command failed")
		} else if strings.Contains(string(output), "build failed") {
			// 编译失败
			fmt.Println("build failed")
		} else {
			// 其他情况
			fmt.Println("failed to run test %s: %v, output %s", testFile, err, string(output))
		}
	}

	// 解析测试结果
	//lines := strings.Split(string(output), "\n")
	//var testResults []TestResult
	//for _, line := range lines {
	//	if strings.HasPrefix(line, "=== RUN") {
	//		parts := strings.Split(line, " ")
	//		testName := parts[2]
	//		testResults = append(testResults, TestResult{
	//			Package: filepath.Dir(testFile),
	//			Test:    testName,
	//			Output:  "",
	//		})
	//	} else if len(testResults) > 0 {
	//		testResults[len(testResults)-1].Output += line + "\n"
	//	}
	//}

	return err, testInfo
}

// getBaseName 获取文件或目录名的基本名称，用于分组
func getBaseName(name string) string {
	ext := filepath.Ext(name)
	base := name[0 : len(name)-len(ext)]

	re := regexp.MustCompile(`(_\d+_test|_test)$`)
	return re.ReplaceAllString(base, "")
}
