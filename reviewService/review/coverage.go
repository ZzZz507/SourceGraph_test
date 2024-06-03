package review

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	constant "reviewService/utils"
)

// GetOneTestFileCover 获取单个离散测试文件的覆盖率
func GetOneTestFileCover(srcFile, testFile string) (error, float32) {
	testFileName := filepath.Base(testFile)
	testFileNameWithoutExt := testFileName[:len(testFileName)-3] // Remove ".go" extension

	// 构建go test命令
	coverProfilePath := filepath.Join(constant.CoverProfileOut, fmt.Sprintf("cover_%s.out", testFileNameWithoutExt))
	testArgs := []string{"test", "-v"}
	testArgs = append(testArgs, srcFile) // 添加file.go到testArgs切片中
	testArgs = append(testArgs, testFile)
	testArgs = append(testArgs, "-cover", "-coverprofile="+coverProfilePath)

	return nil, 0.0
}

// RunGoTestWithCoverages 遍历目录获取所有测试文件的覆盖率情况
func RunGoTestWithCoverages(srcFilePath, testFilePath, coverProfileOut string) error {
	// 在srcFilePath中查找任意名称的源代码文件
	srcFilePattern := ".*\\.go"
	srcFileRegex := regexp.MustCompile(srcFilePattern)
	var srcFiles []string

	err := filepath.Walk(srcFilePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && srcFileRegex.MatchString(info.Name()) {
			srcFiles = append(srcFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to find source files: %v", err)
	}

	// 为每个源文件执行go test -v命令
	for _, srcFile := range srcFiles {
		srcFileName := filepath.Base(srcFile)
		srcFileNameWithoutExt := srcFileName[:len(srcFileName)-3] // Remove ".go" extension
		testFilePattern := fmt.Sprintf("%s_[0-100]_test\\.go", srcFileNameWithoutExt)
		fmt.Println("testFilePattern-----", testFilePattern)
		testFileRegex := regexp.MustCompile(testFilePattern)
		fmt.Println("testFileRegex-----", testFileRegex)
		var testFiles []string

		err := filepath.Walk(testFilePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && testFileRegex.MatchString(info.Name()) {
				testFiles = append(testFiles, path)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to find test files for %s: %v", srcFileName, err)
		}

		// 为每个测试文件执行go test -v命令
		for _, testFile := range testFiles {
			testFileName := filepath.Base(testFile)
			testFileNameWithoutExt := testFileName[:len(testFileName)-3] // Remove ".go" extension

			// 构建go test命令
			coverProfilePath := filepath.Join(coverProfileOut, fmt.Sprintf("cover_%s.out", testFileNameWithoutExt))
			testArgs := []string{"test", "-v"}
			testArgs = append(testArgs, srcFile) // 添加file.go到testArgs切片中
			testArgs = append(testArgs, testFile)
			testArgs = append(testArgs, "-cover", "-coverprofile="+coverProfilePath)

			// 运行测试文件并收集覆盖率信息
			fmt.Println("srcFile", srcFile)
			fmt.Println("测试文件 go", testArgs)
			cmd := exec.Command("go", testArgs...)
			output, err := cmd.CombinedOutput()
			cmd.Dir = "/Users/zhanghunzhuo/Desktop/GolangTestingDemo"
			if err != nil {
				fmt.Printf("Test output for %s:\n%s\n", testFileName, output)
				// return fmt.Errorf("failed to run test with coverage: %v", err)
			}
			fmt.Printf("Test output for %s:\n%s\n", testFileName, output)

			// 分析覆盖率信息并将结果保存到文件中
			coverageOutputFilePath := filepath.Join(coverProfileOut, fmt.Sprintf("coverage_%s.txt", testFileNameWithoutExt))
			coverageOutputFile, err := os.Create(coverageOutputFilePath)
			if err != nil {
				return fmt.Errorf("failed to create coverage output file: %v", err)
			}
			defer coverageOutputFile.Close()

			coverageOutputCmd := exec.Command("go", "tool", "cover", "-func="+coverProfilePath)
			coverageOutputCmd.Stdout = coverageOutputFile
			if err := coverageOutputCmd.Run(); err != nil {
				return fmt.Errorf("failed to get coverage info for %s: %v", testFileName, err)
			}
		}

		//加入总数out
		fmt.Println("testFiles", testFiles)
		if len(testFiles) == 0 {
			fmt.Printf("Skipping file %s as it does not have any tests\n", srcFile)
			continue
		}

		// 构建go test命令
		coverProfilePath := filepath.Join(coverProfileOut, fmt.Sprintf("cover_%s.out", srcFileNameWithoutExt))
		testArgs := []string{"test", "-v"}
		testArgs = append(testArgs, srcFile)
		fmt.Println("go00", testArgs)
		testArgs = append(testArgs, testFiles...)
		fmt.Println("go11", testArgs)
		testArgs = append(testArgs, "-cover", "-coverprofile="+coverProfilePath)

		// 运行测试文件并收集覆盖率信息
		fmt.Println("go22", testArgs)
		cmd := exec.Command("go", testArgs...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Test output for %s:\n%s\n", srcFileName, output)
			// return fmt.Errorf("failed to run test with coverage: %v", err)
		}
		fmt.Printf("Test output for %s:\n%s\n", srcFileName, output)

		// 分析覆盖率信息并将结果保存到文件中
		coverageOutputFilePath := filepath.Join(coverProfileOut, fmt.Sprintf("coverage_%s.txt", srcFileNameWithoutExt))
		coverageOutputFile, err := os.Create(coverageOutputFilePath)
		if err != nil {
			return fmt.Errorf("failed to create coverage output file: %v", err)
		}
		defer coverageOutputFile.Close()

		coverageOutputCmd := exec.Command("go", "tool", "cover", "-func="+coverProfilePath)
		coverageOutputCmd.Stdout = coverageOutputFile
		if err := coverageOutputCmd.Run(); err != nil {
			return fmt.Errorf("failed to get coverage info for %s: %v", srcFileName, err)
		}

	}

	return nil
}
