package loadbalance

type Node struct {
	Weight    int
	Data      interface{}
	effective int
	curWeight int
}

func BuildNodes(weights []int) []*Node {
	nodes := make([]*Node, len(weights))
	for i, w := range weights {
		nodes[i] = &Node{
			Weight: w,
			Data:   i,
		}
	}
	return nodes
}

type NginxScheduler struct {
	nodes []*Node
}

func NewNginxScheduler(nodes []*Node) *NginxScheduler {
	for _, n := range nodes {
		if n.Weight < 0 {
			n.Weight = -n.Weight
			n.effective = n.Weight
		} else if n.Weight == 0 {
			n.effective = 1
		} else {
			n.effective = n.Weight
		}
	}
	return &NginxScheduler{
		nodes: nodes,
	}
}

// 每次调用Next都会遍历所有节点。
// 选出的节点的当前权重会减去所有节点的有效权重之和。
// 对于节点间权重相差比较大的情况，NginxScheduler的选择效果比WeightedScheduler要好一些，更加均衡，但性能要差些。
func (ns *NginxScheduler) Next() *Node {
	total := 0 // total 记录所有节点的有效权重
	var best *Node
	for _, n := range ns.nodes {
		n.curWeight += n.effective // 每次检查都增加当前权重
		total += n.effective

		if n.effective < n.Weight {
			// 节点通了增加权重直到等于weight
			n.effective++
		}
		if best == nil || n.curWeight > best.curWeight {
			// 选择当前权重最大的
			best = n
		}
	}

	if best == nil {
		return nil
	}
	best.curWeight -= total // 被选中的减去total, 这样下次该节点被选中的概率就小了
	return best
}