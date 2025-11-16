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

// GetCommitLog 获取指定范围的 commit 历史
func GetCommitLog(rangeSpec string) (string, error) {
	args := []string{"log", "--pretty=format:%H|%s|%an|%ad", "--date=short"}
	if rangeSpec != "" {
		args = append(args, rangeSpec)
	} else {
		// 如果没有指定范围，获取自上次 tag 以来的 commits
		args = append(args, "--no-merges")
	}
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("无法获取 commit 历史: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetLastTag 获取最后一个 tag
func GetLastTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		// 如果没有 tag，返回空字符串
		return "", nil
	}
	return strings.TrimSpace(string(output)), nil
}

// GetBranchDiff 获取两个分支之间的差异
func GetBranchDiff(baseBranch, targetBranch string) (string, error) {
	cmd := exec.Command("git", "diff", "--stat", fmt.Sprintf("%s..%s", baseBranch, targetBranch))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("无法获取分支差异: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetBranchDiffFull 获取两个分支之间的完整 diff
func GetBranchDiffFull(baseBranch, targetBranch string) (string, error) {
	cmd := exec.Command("git", "diff", fmt.Sprintf("%s..%s", baseBranch, targetBranch))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("无法获取分支完整差异: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetBranchCommits 获取两个分支之间的 commit 列表
func GetBranchCommits(baseBranch, targetBranch string) (string, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%H|%s|%an|%ad", "--date=short", fmt.Sprintf("%s..%s", baseBranch, targetBranch), "--no-merges")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("无法获取分支 commit 列表: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

