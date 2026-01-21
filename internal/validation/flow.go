package validation

import "github.com/pupload/pupload/internal/models"

func flowDetectEmpty(r *ValidationResult, flow models.Flow) {
	isEmpty := len(flow.Nodes) == 0

	if isEmpty {
		r.AddError(ValidationEntry{
			ValidationError,
			ErrFlowEmpty,
			"FlowEmpty",
			"Flow has no nodes, nothing to execute",
		})
	}
}

func flowDetectCycle(r *ValidationResult, flow models.Flow) {
	edgeProducers := make(map[string]string)
	edgeConsumers := make(map[string][]string)

	for _, node := range flow.Nodes {
		for _, output := range node.Outputs {
			edgeProducers[output.Edge] = node.ID
		}

		for _, input := range node.Inputs {
			edgeConsumers[input.Edge] = append(edgeConsumers[input.Edge], node.ID)
		}
	}

	adjacencyList := make(map[string][]string)
	inDegree := make(map[string]int)

	for edgeName, producerID := range edgeProducers {
		consumers := edgeConsumers[edgeName]
		for _, consumerID := range consumers {
			adjacencyList[producerID] = append(adjacencyList[producerID], consumerID)
			inDegree[consumerID]++
		}
	}

	q := make([]string, 0)
	for _, node := range flow.Nodes {
		if inDegree[node.ID] == 0 {
			q = append(q, node.ID)
		}
	}

	processedCount := 0

	for len(q) > 0 {
		current := q[0]
		q = q[1:]
		processedCount++

		for _, successor := range adjacencyList[current] {
			inDegree[successor]--
			if inDegree[successor] == 0 {
				q = append(q, successor)
			}
		}

	}

	isDag := processedCount == len(flow.Nodes)

	if !isDag {
		r.AddError(ValidationEntry{
			ValidationError,
			ErrFlowCycle,
			"FlowCycle",
			"Flow has a cycle, which would cause infinite execution",
		})
	}
}
