package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetStagedDiff 获取暂存区的 diff
func GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("无法获取暂存区变更: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetDiff 获取工作区的 diff
func GetDiff(args ...string) (string, error) {
	cmdArgs := []string{"diff"}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command("git", cmdArgs...)
	output, err := cmd.Output()
	if err != nil {
		// git diff 在无变更时返回非零退出码，这是正常的
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return strings.TrimSpace(string(output)), nil
		}
		return "", fmt.Errorf("无法获取变更: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetFileDiff 获取特定文件的 diff
func GetFileDiff(file string) (string, error) {
	cmd := exec.Command("git", "diff", file)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return strings.TrimSpace(string(output)), nil
		}
		return "", fmt.Errorf("无法获取文件变更: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCommitInfo 获取 commit 的详细信息
func GetCommitInfo(commit string) (string, error) {
	// 获取 commit message
	msgCmd := exec.Command("git", "log", "-1", "--pretty=format:%s%n%n%b", commit)
	msgOutput, err := msgCmd.Output()
	if err != nil {
		return "", fmt.Errorf("无法获取 commit message: %w", err)
	}

	// 获取 commit diff
	diffCmd := exec.Command("git", "show", "--stat", commit)
	diffOutput, err := diffCmd.Output()
	if err != nil {
		return "", fmt.Errorf("无法获取 commit diff: %w", err)
	}

	result := fmt.Sprintf("Commit: %s\n\n", commit)
	result += fmt.Sprintf("Message:\n%s\n\n", string(msgOutput))
	result += fmt.Sprintf("Changes:\n%s", string(diffOutput))

	return result, nil
}

// GetStagedFiles 获取暂存区的文件列表
func GetStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("无法获取暂存区文件列表: %w", err)
	}
	
	if len(output) == 0 {
		return []string{}, nil
	}
	
	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	return files, nil
}

