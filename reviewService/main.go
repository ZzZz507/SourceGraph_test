package main

import (
	"fmt"
	"os"
	"reviewService/review"
	constant "reviewService/utils"
)

func main() {

	err := review.RunGoTestsInProject(constant.RootPath)
	if err != nil {
		fmt.Printf("Error running go tests: %v\n", err)
		os.Exit(1)
	}

	//srcFilePath1 := "/Users/zhanghunzhuo/Desktop/GolangTestingDemo/function/"
	//testFilePath := "/Users/zhanghunzhuo/Desktop/GolangTestingDemo/function/"
	//coverProfileOut := "/Users/zhanghunzhuo/GolandProjects/reviewService/out"
	//err2 := review.RunGoTestWithCoverages(srcFilePath1, testFilePath, coverProfileOut)
	//if err2 != nil {
	//	fmt.Printf("Error running go tests: %v\n", err2)
	//	os.Exit(1)
	//}
}
