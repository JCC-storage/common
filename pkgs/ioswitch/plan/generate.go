package plan

import (
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/dag"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/plan/ops"
)

func Generate(graph *dag.Graph, planBld *exec.PlanBuilder) error {
	generateSend(graph)
	return buildPlan(graph, planBld)
}

// 生成Send指令
func generateSend(graph *dag.Graph) {
	graph.Walk(func(node *dag.Node) bool {
		for _, out := range node.OutputStreams {
			to := out.Toes[0]
			if to.Node.Env.Equals(node.Env) {
				continue
			}

			switch to.Node.Env.Type {
			case dag.EnvDriver:
				// // 如果是要送到Driver，则只能由Driver主动去拉取
				getNode := graph.NewNode(&ops.GetStreamType{}, nil)
				getNode.Env.ToEnvDriver()

				// // 同时需要对此变量生成HoldUntil指令，避免Plan结束时Get指令还未到达
				holdNode := graph.NewNode(&ops.HoldUntilType{}, nil)
				holdNode.Env = node.Env

				// 将Get指令的信号送到Hold指令
				getNode.OutputValues[0].To(holdNode, 0)
				// 将Get指令的输出送到目的地
				getNode.OutputStreams[0].To(to.Node, to.SlotIndex)
				out.Toes = nil
				// 将源节点的输出送到Hold指令
				out.To(holdNode, 0)
				// 将Hold指令的输出送到Get指令
				holdNode.OutputStreams[0].To(getNode, 0)

			case dag.EnvWorker:
				// 如果是要送到Agent，则可以直接发送
				n := graph.NewNode(&ops.SendStreamType{}, nil)
				n.Env = node.Env
				n.OutputStreams[0].To(to.Node, to.SlotIndex)
				out.Toes = nil
				out.To(n, 0)
			}
		}

		for _, out := range node.OutputValues {
			to := out.Toes[0]
			if to.Node.Env.Equals(node.Env) {
				continue
			}

			switch to.Node.Env.Type {
			case dag.EnvDriver:
				// // 如果是要送到Driver，则只能由Driver主动去拉取
				getNode := graph.NewNode(&ops.GetVaType{}, nil)
				getNode.Env.ToEnvDriver()

				// // 同时需要对此变量生成HoldUntil指令，避免Plan结束时Get指令还未到达
				holdNode := graph.NewNode(&ops.HoldUntilType{}, nil)
				holdNode.Env = node.Env

				// 将Get指令的信号送到Hold指令
				getNode.OutputValues[0].To(holdNode, 0)
				// 将Get指令的输出送到目的地
				getNode.OutputValues[1].To(to.Node, to.SlotIndex)
				out.Toes = nil
				// 将源节点的输出送到Hold指令
				out.To(holdNode, 0)
				// 将Hold指令的输出送到Get指令
				holdNode.OutputValues[0].To(getNode, 0)

			case dag.EnvWorker:
				// 如果是要送到Agent，则可以直接发送
				n := graph.NewNode(&ops.SendVarType{}, nil)
				n.Env = node.Env
				n.OutputValues[0].To(to.Node, to.SlotIndex)
				out.Toes = nil
				out.To(n, 0)
			}
		}

		return true
	})
}

// 生成Plan
func buildPlan(graph *dag.Graph, blder *exec.PlanBuilder) error {
	var retErr error
	graph.Walk(func(node *dag.Node) bool {
		for _, out := range node.OutputStreams {
			if out.Var != nil {
				continue
			}

			out.Var = blder.NewStreamVar()
		}

		for _, in := range node.InputStreams {
			if in.Var != nil {
				continue
			}

			in.Var = blder.NewStreamVar()
		}

		for _, out := range node.OutputValues {
			if out.Var != nil {
				continue
			}

			switch out.Type {
			case dag.StringValueVar:
				out.Var = blder.NewStringVar()
			case dag.SignalValueVar:
				out.Var = blder.NewSignalVar()
			}
		}

		for _, in := range node.InputValues {
			if in.Var != nil {
				continue
			}

			switch in.Type {
			case dag.StringValueVar:
				in.Var = blder.NewStringVar()
			case dag.SignalValueVar:
				in.Var = blder.NewSignalVar()
			}
		}

		op, err := node.Type.GenerateOp(node)
		if err != nil {
			retErr = err
			return false
		}

		switch node.Env.Type {
		case dag.EnvDriver:
			blder.AtDriver().AddOp(op)
		case dag.EnvWorker:
			blder.AtWorker(node.Env.Worker).AddOp(op)
		}

		return true
	})

	return retErr
}
