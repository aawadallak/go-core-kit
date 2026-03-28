package idem_test

import (
	"testing"

	"github.com/aawadallak/go-core-kit/core/idem"
)

func TestBuildOperationKey(t *testing.T) {
	t.Parallel()

	key := idem.BuildOperationKey(
		idem.WithAction("donation_paid"),
		idem.WithEntityID("1498"),
		idem.WithCorrelationID("abc-123"),
	)
	if key != "donation_paid:1498:abc-123" {
		t.Fatalf("unexpected key: %s", key)
	}
}

func TestBuildOperationKey_Normalizes(t *testing.T) {
	t.Parallel()

	key := idem.BuildOperationKey(
		idem.WithAction(" Withdrawal Execute Pix "),
		idem.WithEntityID(" tx:1001 "),
		idem.WithCorrelationID(" corr:xyz "),
	)
	if key != "withdrawal_execute_pix:tx_1001:corr_xyz" {
		t.Fatalf("unexpected normalized key: %s", key)
	}
}

func TestBuildOperationKey_DefaultsMissingOptions(t *testing.T) {
	t.Parallel()

	key := idem.BuildOperationKey(
		idem.WithAction("withdrawal_request"),
		idem.WithEntityID("quote-1"),
	)
	if key != "withdrawal_request:quote-1:unknown_correlation" {
		t.Fatalf("unexpected key: %s", key)
	}
}
