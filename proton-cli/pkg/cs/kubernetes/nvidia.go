package kubernetes

import (
	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
)

func checkNvidiaRuntimeAviable(e exec.Executor) bool {
	output, err := e.Command("nvidia-smi").Output()
	if err != nil {
		return false
	} else {
		return len(output) > 0
	}
}
