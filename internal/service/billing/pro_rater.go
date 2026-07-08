package billing

import (
	"errors"
	"math"
	"time"
)

// ProRater handles mid-billing-cycle calculations for plan upgrades or downgrades.
type ProRater struct{}

// NewProRater initializes a new instance of the subscription pro-ration engine.
func NewProRater() *ProRater {
	return &ProRater{}
}

// ProrationResult holds the calculated breakdown of a plan alteration.
type ProrationResult struct {
	UnusedCredit float64 `json:"unused_credit"` // Refundable credit from old plan
	NewPlanCost  float64 `json:"new_plan_cost"` // Cost of new plan for remaining period
	AmountDue    float64 `json:"amount_due"`    // Net amount the customer needs to pay now
	CreditChange float64 `json:"credit_change"` // Excess credit to carry forward (if downgrade)
}

// Calculate computes the pro-rated financial delta when switching packages.
// It tracks time precisely down to the second to ensure fair billing accuracy.
func (p *ProRater) Calculate(
	oldPrice float64,
	newPrice float64,
	cycleStart time.Time,
	cycleEnd time.Time,
) (ProrationResult, error) {

	now := time.Now()

	// 1. Boundary Guard Rails
	if cycleEnd.Before(cycleStart) || cycleEnd.Equal(cycleStart) {
		return ProrationResult{}, errors.New("invalid billing cycle intervals: end date must be after start date")
	}

	// If calculating before the cycle begins, default to the full new plan price.
	if now.Before(cycleStart) {
		return ProrationResult{
			UnusedCredit: oldPrice,
			NewPlanCost:  newPrice,
			AmountDue:    newPrice,
			CreditChange: 0,
		}, nil
	}

	// If calculating after the cycle has elapsed, the cycle is complete.
	if now.After(cycleEnd) {
		return ProrationResult{
			UnusedCredit: 0,
			NewPlanCost:  0,
			AmountDue:    0,
			CreditChange: 0,
		}, nil
	}

	// 2. Precise Time Proportions
	totalCycleDuration := cycleEnd.Sub(cycleStart).Seconds()
	remainingDuration := cycleEnd.Sub(now).Seconds()

	remainingRatio := remainingDuration / totalCycleDuration

	// 3. Financial Delta Generation
	// Calculate what is left of the user's previous payment investment
	unusedCredit := oldPrice * remainingRatio

	// Calculate what the new plan will cost for the remaining period
	newPlanCost := newPrice * remainingRatio

	netDelta := newPlanCost - unusedCredit

	var amountDue float64
	var creditChange float64

	if netDelta >= 0 {
		// Upgrade scenario: Customer owes an additional fee
		amountDue = p.roundToTwoDecimals(netDelta)
		creditChange = 0
	} else {
		// Downgrade scenario: Customer is owed credit to be applied to their wallet/next bill
		amountDue = 0
		creditChange = p.roundToTwoDecimals(math.Abs(netDelta))
	}

	return ProrationResult{
		UnusedCredit: p.roundToTwoDecimals(unusedCredit),
		NewPlanCost:  p.roundToTwoDecimals(newPlanCost),
		AmountDue:    amountDue,
		CreditChange: creditChange,
	}, nil
}

// roundToTwoDecimals ensures currency format accuracy without floating-point overflow drift.
func (p *ProRater) roundToTwoDecimals(val float64) float64 {
	return math.Round(val*100) / 100
}
