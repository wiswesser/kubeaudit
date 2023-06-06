package label

import (
	"github.com/Shopify/kubeaudit"
	"github.com/Shopify/kubeaudit/pkg/k8s"
)

const Name = "label"

const (
	// ImageTagMissing occurs when the deployment label is missing
	LabelMissing = "LabelMissing"
)

// Image implements Auditable
type Label struct {
	value string
}

func New(config Config) *Label {
	return &Label{
		value: config.GetLabel(),
	}
}

// Audit checks that the container image matches the provided image
func (label *Label) Audit(resource k8s.Resource, _ []k8s.Resource) ([]*kubeaudit.AuditResult, error) {
	var auditResults []*kubeaudit.AuditResult

	if resource.GetObjectKind().GroupVersionKind().Kind != "Deployment" {
		return nil, nil
	}

	k8s.GetObjectMeta(resource).GetLabels()

	auditResult := auditDeployment(&resource, label.value)
	if auditResult != nil {
		auditResults = append(auditResults, auditResult)
	}

	return auditResults, nil
}

func auditDeployment(deployment *k8s.Resource, labelValue string) *kubeaudit.AuditResult {
	for k := range k8s.GetObjectMeta(*deployment).GetLabels() {
		if k == labelValue {
			return nil
		}
	}

	return &kubeaudit.AuditResult{
		Auditor:  Name,
		Rule:     LabelMissing,
		Severity: kubeaudit.Warn,
		Message:  "Label is missing.",
	}
}
