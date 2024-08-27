package plan

import (
	"fmt"

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
		switch node.Type.(type) {
		case *ops.SendStreamType:
			return true
		case *ops.SendVarType:
			return true
		case *ops.GetStreamType:
			return true
		case *ops.GetVaType:
			return true
		case *ops.HoldUntilType:
			return true
		}

		for _, out := range node.OutputStreams {
			to := out.Toes[0]
			if to.Node.Env.Equals(node.Env) {
				continue
			}

			switch to.Node.Env.Type {
			case dag.EnvDriver:
				// // 如果是要送到Driver，则只能由Driver主动去拉取
				getNode, getType := dag.NewNode(graph, &ops.GetStreamType{
					FromWorker: node.Env.Worker,
				}, nil)
				getNode.Env.ToEnvDriver()

				// // 同时需要对此变量生成HoldUntil指令，避免Plan结束时Get指令还未到达
				holdNode, holdType := dag.NewNode(graph, &ops.HoldUntilType{}, nil)
				holdNode.Env = node.Env

				// 将Get指令的信号送到Hold指令
				holdType.Signal(holdNode, getType.SignalVar(getNode))

				out.Toes = nil

				// 将源节点的输出送到Hold指令，将Hold指令的输出送到Get指令
				getType.Get(getNode, holdType.HoldStream(holdNode, out)).
					// 将Get指令的输出送到目的地
					To(to.Node, to.SlotIndex)

			case dag.EnvWorker:
				// 如果是要送到Agent，则可以直接发送
				n, t := dag.NewNode(graph, &ops.SendStreamType{
					ToWorker: to.Node.Env.Worker,
				}, nil)
				n.Env = node.Env

				out.Toes = nil
				t.Send(n, out).To(to.Node, to.SlotIndex)
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
				getNode, getType := dag.NewNode(graph, &ops.GetVaType{
					FromWorker: node.Env.Worker,
				}, nil)
				getNode.Env.ToEnvDriver()

				// // 同时需要对此变量生成HoldUntil指令，避免Plan结束时Get指令还未到达
				holdNode, holdType := dag.NewNode(graph, &ops.HoldUntilType{}, nil)
				holdNode.Env = node.Env

				// 将Get指令的信号送到Hold指令
				holdType.Signal(holdNode, getType.SignalVar(getNode))

				out.Toes = nil

				// 将源节点的输出送到Hold指令，将Hold指令的输出送到Get指令
				getType.Get(getNode, holdType.HoldVar(holdNode, out)).
					// 将Get指令的输出送到目的地
					To(to.Node, to.SlotIndex)

			case dag.EnvWorker:
				// 如果是要送到Agent，则可以直接发送
				n, t := dag.NewNode(graph, &ops.SendVarType{
					ToWorker: to.Node.Env.Worker,
				}, nil)
				n.Env = node.Env

				out.Toes = nil
				t.Send(n, out).To(to.Node, to.SlotIndex)
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
			default:
				retErr = fmt.Errorf("unsupported value var type: %v", out.Type)
				return false
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
			default:
				retErr = fmt.Errorf("unsupported value var type: %v", in.Type)
				return false
			}
		}

		op, err := node.Type.GenerateOp(node)
		if err != nil {
			retErr = err
			return false
		}

		// TODO 当前ToDriver，FromDriver不会生成Op，所以这里需要判断一下
		if op == nil {
			return true
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
