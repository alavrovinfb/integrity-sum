package data

import (
	"fmt"
	"strings"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/models"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/api"
)

type HashFile interface {
	PrepareBatchQuery(hd []*api.HashData, dd *models.DeploymentData) (string, []any)
}
type hashFile struct{}

func NewHashFileData() HashFile {
	return &hashFile{}
}

func (o *hashFile) PrepareBatchQuery(hd []*api.HashData, dd *models.DeploymentData) (string, []any) {
	query := `
	INSERT INTO hashfiles (
			file_name,
			full_file_path,
			hash_sum,
			algorithm,
			name_pod,
			image_tag,
			time_of_creation,
			name_deployment
		)
	VALUES `

	fieldsCount := 8
	defaultHashCount := 20
	valuesString := make([]string, 0, defaultHashCount)
	args := make([]any, 0, defaultHashCount*fieldsCount)

	i := 0
	for _, v := range hd {
		valuesString = append(valuesString,
			fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				i*fieldsCount+1, i*fieldsCount+2, i*fieldsCount+3, i*fieldsCount+4,
				i*fieldsCount+5, i*fieldsCount+6, i*fieldsCount+7, i*fieldsCount+8))
		args = append(args,
			v.FullFileName,
			v.Hash,
			v.Algorithm,
			dd.NamePod,
			dd.Image,
			dd.Timestamp,
			dd.NameDeployment,
		)
		i++
	}
	query += strings.Join(valuesString, ", ")
	return query, args
}
