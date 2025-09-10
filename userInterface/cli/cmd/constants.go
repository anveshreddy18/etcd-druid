package cmd

type OperationType string

const (
	OperationReconcile                 OperationType = "Reconcile"
	OperationAddComponentProtection    OperationType = "AddComponentProtection"
	OperationRemoveComponentProtection OperationType = "RemoveComponentProtection"
)
