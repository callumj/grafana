package store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/grafana/grafana/pkg/services/ngalert/models"
	"github.com/grafana/grafana/pkg/services/ngalert/tests"
	"github.com/stretchr/testify/require"
)

const testAlertingIntervalSeconds = 10

func TestProvisioningStore(t *testing.T) {
	_, dbstore := tests.SetupTestEnv(t, testAlertingIntervalSeconds)

	t.Run("Default provenance of a known type is None", func(t *testing.T) {
		rule := models.AlertRule{
			UID: "asdf",
		}

		provenance, err := dbstore.GetProvenance(context.Background(), &rule)

		require.NoError(t, err)
		require.Equal(t, models.ProvenanceNone, provenance)
	})

	t.Run("Store returns saved provenance type", func(t *testing.T) {
		rule := models.AlertRule{
			UID: "123",
		}
		err := dbstore.SetProvenance(context.Background(), &rule, models.ProvenanceFile)
		require.NoError(t, err)

		p, err := dbstore.GetProvenance(context.Background(), &rule)

		require.NoError(t, err)
		require.Equal(t, models.ProvenanceFile, p)
	})

	t.Run("Store does not get provenance of record with different org ID", func(t *testing.T) {
		ruleOrg2 := models.AlertRule{
			UID:   "456",
			OrgID: 2,
		}
		ruleOrg3 := models.AlertRule{
			UID:   "456",
			OrgID: 3,
		}
		err := dbstore.SetProvenance(context.Background(), &ruleOrg2, models.ProvenanceFile)
		require.NoError(t, err)

		p, err := dbstore.GetProvenance(context.Background(), &ruleOrg3)

		require.NoError(t, err)
		require.Equal(t, models.ProvenanceNone, p)
	})

	t.Run("Store only updates provenance of record with given org ID", func(t *testing.T) {
		ruleOrg2 := models.AlertRule{
			UID:   "789",
			OrgID: 2,
		}
		ruleOrg3 := models.AlertRule{
			UID:   "789",
			OrgID: 3,
		}
		err := dbstore.SetProvenance(context.Background(), &ruleOrg2, models.ProvenanceFile)
		require.NoError(t, err)
		err = dbstore.SetProvenance(context.Background(), &ruleOrg3, models.ProvenanceFile)
		require.NoError(t, err)

		err = dbstore.SetProvenance(context.Background(), &ruleOrg2, models.ProvenanceApi)
		require.NoError(t, err)

		p, err := dbstore.GetProvenance(context.Background(), &ruleOrg2)
		require.NoError(t, err)
		require.Equal(t, models.ProvenanceApi, p)
		p, err = dbstore.GetProvenance(context.Background(), &ruleOrg3)
		require.NoError(t, err)
		require.Equal(t, models.ProvenanceFile, p)
	})

	t.Run("Store saves provenance type when transaction is applied", func(t *testing.T) {
		rule := models.AlertRule{
			UID: "456",
		}

		err := dbstore.InTransaction(context.Background(), func(ctx context.Context) error {
			return dbstore.SetProvenance(ctx, &rule, models.ProvenanceFile)
		})
		require.NoError(t, err)

		provenance, err := dbstore.GetProvenance(context.Background(), &rule)
		require.NoError(t, err)
		require.Equal(t, models.ProvenanceFile, provenance)
	})

	t.Run("Transactional store which errors before saving rolls back type update", func(t *testing.T) {
		rule := models.AlertRule{
			UID: "789",
		}

		_ = dbstore.InTransaction(context.Background(), func(ctx context.Context) error {
			err := dbstore.SetProvenance(ctx, &rule, models.ProvenanceFile)
			require.NoError(t, err)
			return fmt.Errorf("something happened!")
		})

		provenance, err := dbstore.GetProvenance(context.Background(), &rule)
		require.NoError(t, err)
		require.Equal(t, models.ProvenanceNone, provenance)
	})
}
