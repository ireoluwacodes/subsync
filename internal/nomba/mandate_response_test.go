package nomba

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPErrorFromNombaBody_ErrorsArray(t *testing.T) {
	raw := []byte(`{"errors":["customerAddress must not be blank"]}`)
	err := HTTPErrorFromNombaBody(200, raw)
	require.Equal(t, "customerAddress must not be blank", err.Error())
	require.Equal(t, []string{"customerAddress must not be blank"}, err.Errors)
}

func TestHTTPErrorFromNombaBody_LegacyMessage(t *testing.T) {
	raw := []byte(`{"responseCode":"01","responseMessage":"Invalid bank code"}`)
	err := HTTPErrorFromNombaBody(200, raw)
	require.Equal(t, "Invalid bank code", err.Error())
	require.Equal(t, "01", err.Code)
}

func TestParseMandateCreateResponse_SuccessLegacy(t *testing.T) {
	raw := json.RawMessage(`{"responseCode":"00","responseMessage":"SUCCESS","data":{"mandateId":"m-1","description":"pay N50"}}`)
	result, err := parseMandateCreateResponse(raw)
	require.NoError(t, err)
	require.Equal(t, "m-1", result.MandateID)
}

func TestParseMandateCreateResponse_SuccessStandard(t *testing.T) {
	raw := json.RawMessage(`{"code":"00","description":"SUCCESS","data":{"mandateId":"m-2","description":"ok"}}`)
	result, err := parseMandateCreateResponse(raw)
	require.NoError(t, err)
	require.Equal(t, "m-2", result.MandateID)
}

func TestParseMandateCreateResponse_ValidationErrors(t *testing.T) {
	raw := json.RawMessage(`{"errors":["startDate must be a date in the present or in the future"]}`)
	_, err := parseMandateCreateResponse(raw)
	require.Error(t, err)
	require.Contains(t, err.Error(), "startDate")
}

func TestHTTPErrorFromNombaBody_CodeAndMessage(t *testing.T) {
	raw := []byte(`{"code":"400","message":"Variable frequency mandates cannot have an amount set","status":false}`)
	err := HTTPErrorFromNombaBody(200, raw)
	require.Equal(t, "Variable frequency mandates cannot have an amount set", err.Error())
}

func TestParseMandateStatusResponse_Success(t *testing.T) {
	raw := json.RawMessage(`{"code":"00","description":"SUCCESS","data":{"mandateId":"m-1","mandateStatus":"Active","mandateAdviceStatus":"Advice sent","customerAccountName":"Jane","customerAccountNumber":"0123456789"}}`)
	result, err := parseMandateStatusResponse(raw)
	require.NoError(t, err)
	require.Equal(t, "Active", result.MandateStatus)
	require.Equal(t, "Advice sent", result.MandateAdviceStatus)
	require.True(t, result.MandateReadyForDebit())
	require.Equal(t, "ready", result.MandateSetupPhase())
}

func TestMandateStatusResult_ActiveAdviceNotSent(t *testing.T) {
	result := MandateStatusResult{
		MandateStatus:       MandateStatusActive,
		MandateAdviceStatus: MandateAdviceNotSent,
	}
	require.False(t, result.MandateReadyForDebit())
	require.Equal(t, "bank_advice", result.MandateSetupPhase())
}

func TestMandateStatusResult_ActiveAdviceSent(t *testing.T) {
	result := MandateStatusResult{
		MandateStatus:       MandateStatusActive,
		MandateAdviceStatus: MandateAdviceSent,
	}
	require.True(t, result.MandateReadyForDebit())
	require.Equal(t, "ready", result.MandateSetupPhase())
}

func TestParseMandateStatusResponse_ActiveAdviceNotSent(t *testing.T) {
	raw := json.RawMessage(`{"code":"00","description":"SUCCESS","data":{"mandateId":"m-1","mandateStatus":"Active","mandateAdviceStatus":"ADVICE_NOT_SENT","customerAccountName":"Jane","customerAccountNumber":"0123456789"}}`)
	result, err := parseMandateStatusResponse(raw)
	require.NoError(t, err)
	require.Equal(t, "ADVICE_NOT_SENT", result.MandateAdviceStatus)
	require.False(t, result.MandateReadyForDebit())
	require.Equal(t, "bank_advice", result.MandateSetupPhase())
}
