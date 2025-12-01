package node

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"pupload/internal/models"
// 	"time"
// )

// type NodeExec struct {
// 	Node    models.Node
// 	NodeDef models.NodeDef
// }

// type StoreLoader interface {
// }

// type RunLoader interface {

// }

// func (ne NodeExec) HandleExecuteNode(rl Run, sl StoreLoader) {

// 	node := f.GetNode(run.FlowName, nodeIndex)
// 	flow, _ := f.GetFlow(run.FlowName)

// 	nodeDef, exists := f.NodeDefs[node.DefName]
// 	if !exists {
// 		log.Fatalf("Attempting to run node that does not have definition")
// 	}

// 	inputs := make(map[string]string)
// 	for _, edge := range node.Inputs {
// 		artifact := run.Artifacts[edge.Edge]

// 		store, _ := f.GetStore(run.FlowName, artifact.StoreName)
// 		url, err := store.GetURL(context.TODO(), artifact.ObjectName, 1*time.Hour)
// 		if err != nil {
// 			fmt.Println(err.Error())
// 		}

// 		inputs[edge.Name] = url.String()
// 	}

// 	outputs := make(map[string]string)
// 	for _, edge := range node.Outputs {

// 		artifact := Artifact{
// 			StoreName:  *flow.DefaultStore,
// 			ObjectName: fmt.Sprintf("%s-%s", edge.Edge, run.ID),
// 			EdgeName:   edge.Edge,
// 		}

// 		if flow.DefaultStore == nil {
// 			f.log.Error("default flow store is nil")
// 		}

// 		store, _ := f.GetStore(run.FlowName, *flow.DefaultStore)

// 		url, err := store.PutURL(context.TODO(), artifact.ObjectName, 10*time.Second)
// 		if err != nil {
// 			f.log.Error("could not generate put url", "err", err)
// 		}

// 		outputs[edge.Name] = url.String()
// 		WaitingURL := WaitingURL{
// 			Artifact: artifact,
// 			PutURL:   url.String(),
// 			TTL:      time.Now().Add(10 * time.Second),
// 		}

// 		run.WaitingURLs = append(run.WaitingURLs, WaitingURL)
// 	}

// 	p := models.NodeExecutePayload{
// 		RunID:      run.ID,
// 		NodeDef:    nodeDef,
// 		Node:       node,
// 		InputURLs:  inputs,
// 		OutputURLs: outputs,
// 	}

// 	info, err := f.executeNode(p)
// 	if err != nil {
// 		f.log.Error("error enqueing node to execute", "err", err, "node", nodeIndex)
// 	}

// 	run.RunningNodes[nodeIndex] = info
// }
