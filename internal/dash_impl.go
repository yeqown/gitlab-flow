package internal

import "github.com/yeqown/gitlab-flow/internal/repository"

type dashImpl struct {
	repo repository.IFlowRepository
}

func NewDash(confPath string, debug bool) IDash {
	return dashImpl{}
}
func (d dashImpl) FeatureDetail(featureBranchName string) ([]byte, error) {
	panic("implement me")
}

func (d dashImpl) MilestoneOverview(milestoneName, mergeRequestURLs string) ([]byte, error) {
	panic("implement me")
}

func (d dashImpl) ProjectDetail(openWeb bool) ([]byte, error) {
	panic("implement me")
}
