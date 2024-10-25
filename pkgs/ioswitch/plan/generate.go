package plan

import (
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/dag"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/exec"
	"gitlink.org.cn/cloudream/common/pkgs/ioswitch/plan/ops"
)

func Generate(graph *dag.Graph, planBld *exec.PlanBuilder) error {
	myGraph := &ops.GraphNodeBuilder{graph}
	generateSend(myGraph)
	return buildPlan(graph, planBld)
}

// 生成Send指令
func generateSend(graph *ops.GraphNodeBuilder) {
	graph.Walk(func(node dag.Node) bool {
		switch node.(type) {
		case *ops.SendStreamNode:
			return true
		case *ops.SendValueNode:
			return true
		case *ops.GetStreamNode:
			return true
		case *ops.GetValueNode:
			return true
		case *ops.HoldUntilNode:
			return true
		}

		for i := 0; i < node.OutputStreams().Len(); i++ {
			out := node.OutputStreams().Get(i)
			to := out.To().Get(0)
			if to.Node.Env().Equals(node.Env()) {
				continue
			}

			switch to.Node.Env().Type {
			case dag.EnvDriver:

				// // 如果是要送到Driver，则只能由Driver主动去拉取
				getNode := graph.NewGetStream(node.Env().Worker)
				getNode.Env().ToEnvDriver()

				// // 同时需要对此变量生成HoldUntil指令，避免Plan结束时Get指令还未到达
				holdType := graph.NewHoldUntil() //dag.NewNode(graph, &ops.HoldUntilNode{}, nil)
				*holdType.Env() = *node.Env()

				// 将Get指令的信号送到Hold指令
				holdType.SetSignal(getNode.SignalVar())

				out.To().RemoveAt(0)

				// 将源节点的输出送到Hold指令，将Hold指令的输出送到Get指令
				getNode.Get(holdType.HoldStream(out)).
					// 将Get指令的输出送到目的地
					Connect(to.Node, to.SlotIndex)

			case dag.EnvWorker:
				// 如果是要送到Agent，则可以直接发送
				n := graph.NewSendStream(to.Node.Env().Worker)
				*n.Env() = *node.Env()

				out.To().RemoveAt(0)
				n.Send(out).Connect(to.Node, to.SlotIndex)
			}
		}

		for i := 0; i < node.OutputValues().Len(); i++ {
			out := node.OutputValues().Get(i)
			// 允许Value变量不被使用
			if out.To().Len() == 0 {
				continue
			}

			to := out.To().Get(0)
			if to.Node.Env().Equals(node.Env()) {
				continue
			}

			switch to.Node.Env().Type {
			case dag.EnvDriver:
				// // 如果是要送到Driver，则只能由Driver主动去拉取
				getNode := graph.NewGetValue(node.Env().Worker)
				getNode.Env().ToEnvDriver()

				// // 同时需要对此变量生成HoldUntil指令，避免Plan结束时Get指令还未到达
				holdNode := graph.NewHoldUntil()
				*holdNode.Env() = *node.Env()

				// 将Get指令的信号送到Hold指令
				holdNode.SetSignal(getNode.SignalVar())

				out.To().RemoveAt(0)

				// 将源节点的输出送到Hold指令，将Hold指令的输出送到Get指令
				getNode.Get(holdNode.HoldVar(out)).
					// 将Get指令的输出送到目的地
					Connect(to.Node, to.SlotIndex)

			case dag.EnvWorker:
				// 如果是要送到Agent，则可以直接发送
				t := graph.NewSendValue(to.Node.Env().Worker)
				*t.Env() = *node.Env()

				out.To().RemoveAt(0)

				t.Send(out).Connect(to.Node, to.SlotIndex)
			}
		}

		return true
	})
}

// 生成Plan
func buildPlan(graph *dag.Graph, blder *exec.PlanBuilder) error {
	var retErr error
	graph.Walk(func(node dag.Node) bool {
		for i := 0; i < node.OutputStreams().Len(); i++ {
			out := node.OutputStreams().Get(i)

			if out.VarID > 0 {
				continue
			}

			out.VarID = blder.NewVar()
		}

		for i := 0; i < node.InputStreams().Len(); i++ {
			in := node.InputStreams().Get(i)

			if in.VarID > 0 {
				continue
			}

			in.VarID = blder.NewVar()
		}

		for i := 0; i < node.OutputValues().Len(); i++ {
			out := node.OutputValues().Get(i)

			if out.VarID > 0 {
				continue
			}

			out.VarID = blder.NewVar()
		}

		for i := 0; i < node.InputValues().Len(); i++ {
			in := node.InputValues().Get(i)

			if in.VarID > 0 {
				continue
			}

			in.VarID = blder.NewVar()
		}

		op, err := node.GenerateOp()
		if err != nil {
			retErr = err
			return false
		}

		// TODO 当前ToDriver，FromDriver不会生成Op，所以这里需要判断一下
		if op == nil {
			return true
		}

		switch node.Env().Type {
		case dag.EnvDriver:
			blder.AtDriver().AddOp(op)
		case dag.EnvWorker:
			blder.AtWorker(node.Env().Worker).AddOp(op)
		}

		return true
	})

	return retErr
}
