package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*AllocationCreateRequest)(nil)
	_ twapi.HTTPResponser = (*AllocationCreateResponse)(nil)
	_ twapi.HTTPRequester = (*AllocationUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*AllocationUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*AllocationDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*AllocationDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*AllocationRestoreRequest)(nil)
	_ twapi.HTTPResponser = (*AllocationRestoreResponse)(nil)
	_ twapi.HTTPRequester = (*AllocationTaskLinkRequest)(nil)
	_ twapi.HTTPResponser = (*AllocationTaskLinkResponse)(nil)
	_ twapi.HTTPRequester = (*AllocationTaskUnlinkRequest)(nil)
	_ twapi.HTTPResponser = (*AllocationTaskUnlinkResponse)(nil)
	_ twapi.HTTPRequester = (*AllocationGetRequest)(nil)
	_ twapi.HTTPResponser = (*AllocationGetResponse)(nil)
	_ twapi.HTTPRequester = (*AllocationListRequest)(nil)
	_ twapi.HTTPResponser = (*AllocationListResponse)(nil)
)

// AllocationStatus identifies the lifecycle state of an allocation. Deleting an
// allocation is a soft delete, so a deleted allocation keeps its row and can be
// restored; see AllocationRestore.
type AllocationStatus string

// Supported allocation statuses.
const (
	AllocationStatusActive  AllocationStatus = "active"
	AllocationStatusDeleted AllocationStatus = "deleted"
)

// AllocationSideload identifies a related entity that can be sideloaded into an
// allocation response under the "included" key.
type AllocationSideload string

// Supported allocation sideloads.
const (
	AllocationSideloadProjects          AllocationSideload = "projects"
	AllocationSideloadProjectsCompanies AllocationSideload = "projects.companies"
	AllocationSideloadAssignee          AllocationSideload = "assignee"
	AllocationSideloadAssigneeTasks     AllocationSideload = "assignee.tasks"
	AllocationSideloadAssigneeJobRoles  AllocationSideload = "assignee.jobRoles"

	// AllocationSideloadFinancialDetails populates each allocation's
	// FinancialDetails. It is gated by both an account entitlement and a
	// per-project financial permission, and is dropped from the request rather
	// than rejected when either is missing — read CanViewFinancialDetails on the
	// row to tell an empty result from a withheld one.
	AllocationSideloadFinancialDetails AllocationSideload = "financialDetails"
)

// AllocationProjectStatus filters allocations by the status of the project they
// belong to.
type AllocationProjectStatus string

// Supported project statuses for the allocation list filters.
const (
	AllocationProjectStatusActive    AllocationProjectStatus = "active"
	AllocationProjectStatusCurrent   AllocationProjectStatus = "current"
	AllocationProjectStatusLate      AllocationProjectStatus = "late"
	AllocationProjectStatusUpcoming  AllocationProjectStatus = "upcoming"
	AllocationProjectStatusCompleted AllocationProjectStatus = "completed"
	AllocationProjectStatusDeleted   AllocationProjectStatus = "deleted"
)

// AllocationBillableRateDetails is one effective billable-rate period within an
// allocation. Exactly one of UserEffectiveRate or PlaceholderUserEffectiveRate
// is set: the first when the rate comes from the user's own effective rate, the
// second when it is sourced from the job role of a placeholder user.
type AllocationBillableRateDetails struct {
	// UserEffectiveRate identifies the user effective rate the figures came from.
	UserEffectiveRate *twapi.Relationship `json:"userEffectiveRate,omitempty"`

	// PlaceholderUserEffectiveRate identifies the job-role billable rate the
	// figures came from, for a placeholder user.
	PlaceholderUserEffectiveRate *twapi.Relationship `json:"placeholderUserEffectiveRate,omitempty"`

	// JobRole identifies the job role the rate belongs to, when applicable.
	JobRole *twapi.Relationship `json:"jobRoleId,omitempty"`

	// BillableRate is the rate applied over this period.
	BillableRate float64 `json:"billableRate"`

	// AllocatedMinutes is the allocated time this period covers.
	AllocatedMinutes int64 `json:"allocatedMinutes"`

	// ForecastedRevenue is the revenue the allocated time would produce at this
	// rate if every allocated hour were logged and billed.
	ForecastedRevenue float64 `json:"forecastedRevenue"`

	// Source names where the rate was derived from.
	Source *EffectiveRateSource `json:"source"`

	// EffectiveDate is the date the rate took effect.
	EffectiveDate twapi.Date `json:"effectiveDate"`

	// Currency is the currency the rate is expressed in.
	Currency twapi.Relationship `json:"currency"`
}

// AllocationCostRateDetails is one cost-rate period within an allocation.
// Exactly one of UserCostRate or JobRoleCostRate is set, following the same
// user-versus-placeholder split as AllocationBillableRateDetails.
type AllocationCostRateDetails struct {
	// UserCostRate identifies the user cost rate the figures came from.
	UserCostRate *twapi.Relationship `json:"userCostRate,omitempty"`

	// JobRoleCostRate identifies the job-role cost rate the figures came from,
	// for a placeholder user.
	JobRoleCostRate *twapi.Relationship `json:"jobRoleCostRate,omitempty"`

	// CostRate is the rate applied over this period.
	CostRate float64 `json:"costRate"`

	// AllocatedMinutes is the allocated time this period covers.
	AllocatedMinutes int64 `json:"allocatedMinutes"`

	// ForecastedCost is the cost the allocated time would produce at this rate.
	ForecastedCost float64 `json:"forecastedCost"`

	// Currency is the currency the rate is expressed in.
	Currency twapi.Relationship `json:"currency"`

	// Source names where the rate was derived from.
	Source *CostRateSource `json:"source,omitempty"`
}

// AllocationFinancialDetails carries the money an allocation is forecast to be
// worth. It is only populated when AllocationSideloadFinancialDetails is
// requested and the caller is entitled to see it.
type AllocationFinancialDetails struct {
	// ForecastedRevenue is what the allocated time would be worth in revenue if
	// every allocated hour were logged and billed.
	ForecastedRevenue *float64 `json:"forecastedRevenue,omitempty"`

	// ForecastedCost is the same calculation against cost rates.
	ForecastedCost *float64 `json:"forecastedCost,omitempty"`

	// BillableRates breaks the forecasted revenue down by rate period.
	BillableRates []AllocationBillableRateDetails `json:"billableRates,omitempty"`

	// CostRates breaks the forecasted cost down by rate period.
	CostRates []AllocationCostRateDetails `json:"costRates,omitempty"`
}

// Allocation commits a user's time to a project over a date range. It is the
// core entity of the resource scheduler, and a separate plane of planning from
// tasks: an allocation carries its own committed time, and is not composed of
// the estimates of the tasks linked to it.
//
// What is held constant is the per-day rate, so the time an allocation comes to
// is the working days in its range multiplied by that rate: widening the range
// adds time rather than spreading the same total more thinly. Read the rate from
// SecondsPerDay rather than HoursPerDay — the two describe the same quantity,
// but the hours figure is a float and rounds.
//
// More information can be found at:
// https://support.teamwork.com/projects/schedule/schedule-introduction
//
// sparsefields:gen
type Allocation struct {
	// ID is the unique identifier of the allocation.
	ID int64 `json:"id"`

	// Project is the project the time is committed to.
	Project twapi.Relationship `json:"project"`

	// AssignedUser is the user whose time is committed. It may be a real person
	// or a placeholder user, and nothing else in the response distinguishes the
	// two.
	AssignedUser twapi.Relationship `json:"assignedUser"`

	// Title is the name of the allocation.
	Title string `json:"title"`

	// Description is the optional description of the allocation.
	Description *string `json:"description"`

	// StartDate is the first day of the allocation's range.
	StartDate twapi.Date `json:"startedAt"`

	// EndDate is the last day of the allocation's range.
	EndDate twapi.Date `json:"endedAt"`

	// Duration is the committed time in minutes. It is stored alongside the
	// per-day rate and kept in step with it, but the rate is what the arithmetic
	// reads.
	Duration int64 `json:"duration"`

	// AvailableDuration is AllocatedDuration further reduced by the all-day
	// unavailable events covering the range — time off and the like. In minutes,
	// and floored at zero.
	//
	// Unavailable time is the only thing separating it from AllocatedDuration;
	// both already have non-working days taken out.
	AvailableDuration *int64 `json:"availableDuration,omitempty"`

	// AllocatedDuration is the per-day rate applied across the days of the
	// range, with the days the user has no working hours for removed. In
	// minutes, and floored at zero.
	//
	// Despite the name, neither this nor AvailableDuration says anything about
	// the tasks linked to the allocation, so neither is a measure of how much of
	// the commitment is accounted for.
	AllocatedDuration *int64 `json:"allocatedDuration,omitempty"`

	// HoursPerDay is the time placed on each working day in the range, expressed
	// in hours. Prefer SecondsPerDay: this is the same quantity as a float, so
	// it rounds, and a rate that is not a whole number of hours does not survive
	// it exactly.
	HoursPerDay float64 `json:"hoursPerDay"`

	// SecondsPerDay is the time placed on each working day in the range,
	// expressed in seconds. This is the per-day rate to read: it is exact, where
	// HoursPerDay is a rounded float over the same value.
	SecondsPerDay int64 `json:"secondsPerDay"`

	// Color is the allocation's colour as six hexadecimal digits. The API
	// returns it without a leading "#", which is why this is a plain string
	// rather than a twapi.HexColor: that type requires the "#" and would reject
	// every value the API sends.
	Color string `json:"color"`

	// Status is the lifecycle state of the allocation. A deleted allocation is
	// only returned when the list request asks for deleted rows.
	Status AllocationStatus `json:"status"`

	// IsBillable reports whether the allocated time can be charged to a client.
	IsBillable bool `json:"isBillable"`

	// OverAllocated reports whether this allocation puts the user over their
	// capacity for the period.
	OverAllocated bool `json:"overAllocated"`

	// RecurringRule is the recurrence rule of a recurring allocation, if any.
	// The stored allocation is the seed; the occurrences in a viewed range are
	// expanded on read and are not persisted.
	RecurringRule *string `json:"recurringRule"`

	// LinkedTasks are the tasks associated with the allocation. The
	// association is many-to-many, and the allocation acts as an envelope: the
	// allocation's time is not composed of the tasks' estimates, and removing
	// the allocation drops the link rather than the task.
	LinkedTasks []twapi.Relationship `json:"linkedTaskIDs"`

	// LinkedTaskEstimatedTime is the sum of the linked tasks' estimates, in
	// minutes.
	//
	// This counts each linked task whole, and a task can sit behind more than
	// one allocation, so the figure must not be summed across allocations — it
	// is a per-allocation view, not an additive one.
	LinkedTaskEstimatedTime *int64 `json:"linkedTaskEstimatedTime"`

	// FinancialDetails carries the forecasted revenue and cost of the allocated
	// time. Only populated when the financialDetails sideload is requested and
	// permitted.
	FinancialDetails AllocationFinancialDetails `json:"financialDetails"`

	// CanViewFinancialDetails reports whether the caller is allowed to see
	// FinancialDetails. When false, an empty FinancialDetails means "withheld"
	// rather than "nothing to report".
	CanViewFinancialDetails bool `json:"canViewFinancialDetails"`

	// CreatedAt is the date and time the allocation was created.
	CreatedAt time.Time `json:"createdAt"`

	// CreatedBy is the identifier of the user who created the allocation.
	CreatedBy int64 `json:"createdBy"`

	// UpdatedAt is the date and time the allocation was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`

	// UpdatedBy is the identifier of the user who last updated the allocation.
	UpdatedBy *int64 `json:"updatedBy"`

	// DeletedAt is the date and time the allocation was deleted, if it was.
	DeletedAt *time.Time `json:"deletedAt"`

	// DeletedBy is the identifier of the user who deleted the allocation, if it
	// was deleted.
	DeletedBy *int64 `json:"deletedBy"`
}

// AllocationUpsert carries the writable attributes of an allocation. It is the
// body both the create and the update endpoints accept, under the "allocation"
// key.
type AllocationUpsert struct {
	// ProjectID is the project to commit the time to. Required on create.
	ProjectID *int64 `json:"projectId,omitempty"`

	// AssignedUserID is the user whose time is committed. It accepts a real
	// person or a placeholder user. Required on create.
	AssignedUserID *int64 `json:"assignedUserID,omitempty"`

	// Title is the name of the allocation, at most 100 characters. Required on
	// create.
	Title *string `json:"title,omitempty"`

	// Description is an optional description, at most 255 characters.
	Description *string `json:"description,omitempty"`

	// StartDate is the first day of the range. Required on create.
	StartDate *twapi.Date `json:"startedAt,omitempty"`

	// EndDate is the last day of the range, and must not precede StartDate.
	// Required on create.
	EndDate *twapi.Date `json:"endedAt,omitempty"`

	// SecondsPerDay is the time to place on each working day of the range,
	// expressed in seconds. This is the preferred way to set the per-day rate:
	// it is exact, where HoursPerDay is a float over the same value and rounds.
	// It must come to between one minute and twenty-four hours. The rate is what
	// is held constant, so widening the range adds time rather than spreading
	// the same total.
	SecondsPerDay *int64 `json:"secondsPerDay,omitempty"`

	// HoursPerDay sets the same per-day rate as SecondsPerDay, in hours. Set one
	// or the other, not both; prefer SecondsPerDay unless the rate is a whole
	// number of hours.
	HoursPerDay *float64 `json:"hoursPerDay,omitempty"`

	// Duration is the total committed time in minutes. Prefer SecondsPerDay: when
	// only Duration is sent the API derives a daily rate from it and clamps that
	// rate into the permitted range, which can change the total that is stored.
	Duration *int64 `json:"duration,omitempty"`

	// Color is the allocation's colour as six hexadecimal digits, with or
	// without a leading "#". Required on create.
	Color *string `json:"color,omitempty"`

	// IsBillable marks the allocated time as chargeable to a client.
	IsBillable *bool `json:"isBillable,omitempty"`

	// IgnoreCollisions bypasses the error raised when the range overlaps another
	// allocation for the same user on the same project.
	IgnoreCollisions *bool `json:"ignoreCollisions,omitempty"`

	// InformOfOverAllocation asks the API to report, rather than reject, a change
	// that puts the user over capacity.
	InformOfOverAllocation *bool `json:"informOfOverAllocation,omitempty"`

	// LinkedTaskIDs are the tasks to associate with the allocation. This
	// replaces the whole set: send every task that should remain linked, or use
	// AllocationTaskLink and AllocationTaskUnlink to change the set one task at
	// a time.
	//
	// The key is spelled linkedTaskIDs to match the response and the rest of the
	// API's clients. The endpoint documents the request key as linkedTaskIds and
	// accepts either, since it binds case-insensitively.
	LinkedTaskIDs []int64 `json:"linkedTaskIDs,omitempty"`
}

// AllocationCreateRequest represents the request body for creating a new
// allocation.
type AllocationCreateRequest struct {
	// Allocation carries the attributes of the allocation to create.
	Allocation AllocationUpsert
}

// NewAllocationCreateRequest creates a new AllocationCreateRequest with the
// fields the API requires.
func NewAllocationCreateRequest(
	projectID int64,
	assignedUserID int64,
	title string,
	startDate twapi.Date,
	endDate twapi.Date,
	secondsPerDay int64,
	color string,
) AllocationCreateRequest {
	return AllocationCreateRequest{
		Allocation: AllocationUpsert{
			ProjectID:      &projectID,
			AssignedUserID: &assignedUserID,
			Title:          &title,
			StartDate:      &startDate,
			EndDate:        &endDate,
			SecondsPerDay:  &secondsPerDay,
			Color:          &color,
		},
	}
}

// HTTPRequest creates an HTTP request for the AllocationCreateRequest.
func (a AllocationCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/allocations.json"

	payload := struct {
		Allocation AllocationUpsert `json:"allocation"`
	}{Allocation: a.Allocation}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create allocation request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// AllocationCreateResponse represents the response body for creating a new
// allocation.
type AllocationCreateResponse struct {
	// Allocation is the created allocation.
	Allocation Allocation `json:"allocation"`
}

// HandleHTTPResponse handles the HTTP response for the
// AllocationCreateResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
func (a *AllocationCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create allocation")
	}
	if err := json.NewDecoder(resp.Body).Decode(a); err != nil {
		return fmt.Errorf("failed to decode create allocation response: %w", err)
	}
	if a.Allocation.ID == 0 {
		return fmt.Errorf("create allocation response does not contain a valid identifier")
	}
	return nil
}

// AllocationCreate creates a new allocation using the provided request and
// returns the response.
func AllocationCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req AllocationCreateRequest,
) (*AllocationCreateResponse, error) {
	return twapi.Execute[AllocationCreateRequest, *AllocationCreateResponse](ctx, engine, req)
}

// AllocationUpdateRequestPath contains the path parameters for updating an
// allocation.
type AllocationUpdateRequestPath struct {
	// ID is the unique identifier of the allocation to be updated.
	ID int64
}

// AllocationUpdateRequest represents the request body for updating an
// allocation. Besides the identifier every field is optional, and an omitted
// field is left as it is.
type AllocationUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path AllocationUpdateRequestPath

	// Allocation carries the attributes to change.
	Allocation AllocationUpsert
}

// NewAllocationUpdateRequest creates a new AllocationUpdateRequest with the
// provided allocation ID. The ID is required to update an allocation.
func NewAllocationUpdateRequest(allocationID int64) AllocationUpdateRequest {
	return AllocationUpdateRequest{
		Path: AllocationUpdateRequestPath{
			ID: allocationID,
		},
	}
}

// HTTPRequest creates an HTTP request for the AllocationUpdateRequest.
func (a AllocationUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if a.Path.ID == 0 {
		return nil, fmt.Errorf("an allocation ID is required to update an allocation")
	}
	uri := server + "/projects/api/v3/allocations/" + strconv.FormatInt(a.Path.ID, 10) + ".json"

	payload := struct {
		Allocation AllocationUpsert `json:"allocation"`
	}{Allocation: a.Allocation}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update allocation request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// AllocationUpdateResponse represents the response body for updating an
// allocation.
type AllocationUpdateResponse struct {
	// Allocation is the updated allocation.
	Allocation Allocation `json:"allocation"`
}

// HandleHTTPResponse handles the HTTP response for the
// AllocationUpdateResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
func (a *AllocationUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update allocation")
	}
	if err := json.NewDecoder(resp.Body).Decode(a); err != nil {
		return fmt.Errorf("failed to decode update allocation response: %w", err)
	}
	return nil
}

// AllocationUpdate updates an allocation using the provided request and returns
// the response.
func AllocationUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req AllocationUpdateRequest,
) (*AllocationUpdateResponse, error) {
	return twapi.Execute[AllocationUpdateRequest, *AllocationUpdateResponse](ctx, engine, req)
}

// AllocationDeleteRequestPath contains the path parameters for deleting an
// allocation.
type AllocationDeleteRequestPath struct {
	// ID is the unique identifier of the allocation to be deleted.
	ID int64
}

// AllocationDeleteRequest represents the request body for deleting an
// allocation. The endpoint takes a body, which is unusual for a delete: it is
// what carries HardDelete.
type AllocationDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path AllocationDeleteRequestPath

	// HardDelete removes the allocation permanently. Left false the delete is a
	// soft delete: the allocation is still returned by a list request asking for
	// deleted rows, and AllocationRestore can bring it back.
	HardDelete bool
}

// NewAllocationDeleteRequest creates a new AllocationDeleteRequest with the
// provided allocation ID. The delete is a soft delete unless HardDelete is set.
func NewAllocationDeleteRequest(allocationID int64) AllocationDeleteRequest {
	return AllocationDeleteRequest{
		Path: AllocationDeleteRequestPath{
			ID: allocationID,
		},
	}
}

// HTTPRequest creates an HTTP request for the AllocationDeleteRequest.
func (a AllocationDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if a.Path.ID == 0 {
		return nil, fmt.Errorf("an allocation ID is required to delete an allocation")
	}
	uri := server + "/projects/api/v3/allocations/" + strconv.FormatInt(a.Path.ID, 10) + ".json"

	// An explicit payload, as create and update use, rather than encoding the
	// request itself: the body is then defined by this struct alone, so a field
	// added to AllocationDeleteRequest later cannot silently change the wire
	// shape of a delete.
	payload := struct {
		HardDelete bool `json:"hardDelete"`
	}{HardDelete: a.HardDelete}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode delete allocation request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// AllocationDeleteResponse represents the response body for deleting an
// allocation.
type AllocationDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// AllocationDeleteResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
func (a *AllocationDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete allocation")
	}
	return nil
}

// AllocationDelete deletes an allocation using the provided request and returns
// the response.
func AllocationDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req AllocationDeleteRequest,
) (*AllocationDeleteResponse, error) {
	return twapi.Execute[AllocationDeleteRequest, *AllocationDeleteResponse](ctx, engine, req)
}

// AllocationRestoreRequestPath contains the path parameters for restoring an
// allocation.
type AllocationRestoreRequestPath struct {
	// ID is the unique identifier of the allocation to be restored.
	ID int64
}

// AllocationRestoreRequest represents the request for restoring a soft-deleted
// allocation.
type AllocationRestoreRequest struct {
	// Path contains the path parameters for the request.
	Path AllocationRestoreRequestPath
}

// NewAllocationRestoreRequest creates a new AllocationRestoreRequest with the
// provided allocation ID.
func NewAllocationRestoreRequest(allocationID int64) AllocationRestoreRequest {
	return AllocationRestoreRequest{
		Path: AllocationRestoreRequestPath{
			ID: allocationID,
		},
	}
}

// HTTPRequest creates an HTTP request for the AllocationRestoreRequest.
func (a AllocationRestoreRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if a.Path.ID == 0 {
		return nil, fmt.Errorf("an allocation ID is required to restore an allocation")
	}
	uri := server + "/projects/api/v3/allocations/" + strconv.FormatInt(a.Path.ID, 10) + "/restore.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// AllocationRestoreResponse represents the response body for restoring an
// allocation.
type AllocationRestoreResponse struct {
	// Allocation is the restored allocation.
	Allocation Allocation `json:"allocation"`
}

// HandleHTTPResponse handles the HTTP response for the
// AllocationRestoreResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
func (a *AllocationRestoreResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to restore allocation")
	}
	if err := json.NewDecoder(resp.Body).Decode(a); err != nil {
		return fmt.Errorf("failed to decode restore allocation response: %w", err)
	}
	return nil
}

// AllocationRestore restores a soft-deleted allocation using the provided
// request and returns the response.
func AllocationRestore(
	ctx context.Context,
	engine *twapi.Engine,
	req AllocationRestoreRequest,
) (*AllocationRestoreResponse, error) {
	return twapi.Execute[AllocationRestoreRequest, *AllocationRestoreResponse](ctx, engine, req)
}

// AllocationTaskLinkRequestPath contains the path parameters for linking a task
// to an allocation.
type AllocationTaskLinkRequestPath struct {
	// AllocationID is the unique identifier of the allocation.
	AllocationID int64

	// TaskID is the unique identifier of the task to link.
	TaskID int64
}

// AllocationTaskLinkRequest represents the request for linking one task to an
// allocation. The link is incremental — it leaves the allocation's other links
// alone, unlike AllocationUpsert.LinkedTaskIDs, which replaces the whole set.
// The task and the allocation must belong to the same project.
type AllocationTaskLinkRequest struct {
	// Path contains the path parameters for the request.
	Path AllocationTaskLinkRequestPath
}

// NewAllocationTaskLinkRequest creates a new AllocationTaskLinkRequest for the
// given allocation and task.
func NewAllocationTaskLinkRequest(allocationID, taskID int64) AllocationTaskLinkRequest {
	return AllocationTaskLinkRequest{
		Path: AllocationTaskLinkRequestPath{
			AllocationID: allocationID,
			TaskID:       taskID,
		},
	}
}

// HTTPRequest creates an HTTP request for the AllocationTaskLinkRequest.
func (a AllocationTaskLinkRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if a.Path.AllocationID == 0 || a.Path.TaskID == 0 {
		return nil, fmt.Errorf("an allocation ID and a task ID are required to link a task to an allocation")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/allocations/%d/link/%d.json", server, a.Path.AllocationID, a.Path.TaskID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// AllocationTaskLinkResponse represents the response body for linking a task to
// an allocation.
type AllocationTaskLinkResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// AllocationTaskLinkResponse. If some unexpected HTTP status code is returned
// by the API, a twapi.HTTPError is returned.
func (a *AllocationTaskLinkResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to link task to allocation")
	}
	return nil
}

// AllocationTaskLink links a task to an allocation using the provided request
// and returns the response.
func AllocationTaskLink(
	ctx context.Context,
	engine *twapi.Engine,
	req AllocationTaskLinkRequest,
) (*AllocationTaskLinkResponse, error) {
	return twapi.Execute[AllocationTaskLinkRequest, *AllocationTaskLinkResponse](ctx, engine, req)
}

// AllocationTaskUnlinkRequestPath contains the path parameters for unlinking a
// task from an allocation.
type AllocationTaskUnlinkRequestPath struct {
	// AllocationID is the unique identifier of the allocation.
	AllocationID int64

	// TaskID is the unique identifier of the task to unlink.
	TaskID int64
}

// AllocationTaskUnlinkRequest represents the request for unlinking one task
// from an allocation. It removes the association only: both the task and the
// allocation are left standing.
type AllocationTaskUnlinkRequest struct {
	// Path contains the path parameters for the request.
	Path AllocationTaskUnlinkRequestPath
}

// NewAllocationTaskUnlinkRequest creates a new AllocationTaskUnlinkRequest for
// the given allocation and task.
func NewAllocationTaskUnlinkRequest(allocationID, taskID int64) AllocationTaskUnlinkRequest {
	return AllocationTaskUnlinkRequest{
		Path: AllocationTaskUnlinkRequestPath{
			AllocationID: allocationID,
			TaskID:       taskID,
		},
	}
}

// HTTPRequest creates an HTTP request for the AllocationTaskUnlinkRequest.
func (a AllocationTaskUnlinkRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if a.Path.AllocationID == 0 || a.Path.TaskID == 0 {
		return nil, fmt.Errorf("an allocation ID and a task ID are required to unlink a task from an allocation")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/allocations/%d/unlink/%d.json", server, a.Path.AllocationID, a.Path.TaskID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// AllocationTaskUnlinkResponse represents the response body for unlinking a
// task from an allocation.
type AllocationTaskUnlinkResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// AllocationTaskUnlinkResponse. If some unexpected HTTP status code is returned
// by the API, a twapi.HTTPError is returned.
func (a *AllocationTaskUnlinkResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to unlink task from allocation")
	}
	return nil
}

// AllocationTaskUnlink unlinks a task from an allocation using the provided
// request and returns the response.
func AllocationTaskUnlink(
	ctx context.Context,
	engine *twapi.Engine,
	req AllocationTaskUnlinkRequest,
) (*AllocationTaskUnlinkResponse, error) {
	return twapi.Execute[AllocationTaskUnlinkRequest, *AllocationTaskUnlinkResponse](ctx, engine, req)
}

// AllocationGetRequestPath contains the path parameters for loading a single
// allocation.
type AllocationGetRequestPath struct {
	// ID is the unique identifier of the allocation to be retrieved.
	ID int64
}

// AllocationGetRequest represents the request for loading a single allocation.
type AllocationGetRequest struct {
	// Path contains the path parameters for the request.
	Path AllocationGetRequestPath

	// Include is the list of related entities to sideload into the response.
	Include []AllocationSideload

	// Fields restricts the attributes returned for the allocation. Each slot of
	// AllocationGetFields is a separate `fields[entity]=…` selection; populated
	// slots restrict the response, empty slots return the API default. Use the
	// generated AllocationField constants to ensure values match real
	// attributes.
	Fields AllocationGetFields
}

// NewAllocationGetRequest creates a new AllocationGetRequest with the provided
// allocation ID. The ID is required to load an allocation.
func NewAllocationGetRequest(allocationID int64) AllocationGetRequest {
	return AllocationGetRequest{
		Path: AllocationGetRequestPath{
			ID: allocationID,
		},
	}
}

// HTTPRequest creates an HTTP request for the AllocationGetRequest.
func (a AllocationGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if a.Path.ID == 0 {
		return nil, fmt.Errorf("an allocation ID is required to retrieve an allocation")
	}
	uri := server + "/projects/api/v3/allocations/" + strconv.FormatInt(a.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	querySetStrings(query, "include", a.Include)
	a.Fields.apply(query)
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// AllocationGetResponse contains all the information related to an allocation.
//
// sparsefields:get
type AllocationGetResponse struct {
	Allocation Allocation `json:"allocation"`

	// Included contains the related objects requested through Include.
	//
	// The rate objects the financialDetails sideload also populates —
	// effectiveUserRate and costRate — are deliberately absent: the figures they
	// carry are already summarised inline on the allocation's FinancialDetails,
	// and modelling them would mean two new entities for the detail that is left.
	Included struct {
		// Projects contains the projects the allocations are committed to, keyed
		// by the string representation of the project ID.
		Projects map[string]Project `json:"projects,omitempty"`

		// Companies contains the companies owning those projects, keyed by the
		// string representation of the company ID.
		Companies map[string]Company `json:"companies,omitempty"`

		// Users contains the assigned users, keyed by the string representation of
		// the user ID. A user here may be a placeholder rather than a real person.
		Users map[string]User `json:"users,omitempty"`

		// JobRoles contains the job roles of the assigned users, keyed by the
		// string representation of the job role ID.
		//
		// The response spells this key jobRoles while the endpoint reads the
		// sparse fieldset from fields[jobroles], so the entity name is overridden
		// rather than inherited from the tag.
		//
		// sparsefields:key=jobroles
		JobRoles map[string]JobRole `json:"jobRoles,omitempty"`

		// Tasks contains the assigned users' tasks, keyed by the string
		// representation of the task ID.
		Tasks map[string]Task `json:"tasks,omitempty"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the AllocationGetResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (a *AllocationGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve allocation")
	}
	if err := json.NewDecoder(resp.Body).Decode(a); err != nil {
		return fmt.Errorf("failed to decode retrieve allocation response: %w", err)
	}
	return nil
}

// AllocationGet retrieves a single allocation using the provided request and
// returns the response.
func AllocationGet(
	ctx context.Context,
	engine *twapi.Engine,
	req AllocationGetRequest,
) (*AllocationGetResponse, error) {
	return twapi.Execute[AllocationGetRequest, *AllocationGetResponse](ctx, engine, req)
}

// AllocationOrderBy identifies the attributes an allocation list can be ordered
// by. The values are lowercase and unseparated, which is what the endpoint
// accepts.
type AllocationOrderBy string

// Supported allocation order-by values.
const (
	AllocationOrderByStartDate    AllocationOrderBy = "startdate"
	AllocationOrderByEndDate      AllocationOrderBy = "enddate"
	AllocationOrderByProject      AllocationOrderBy = "project"
	AllocationOrderByAssignedUser AllocationOrderBy = "assigneduser"
	AllocationOrderByID           AllocationOrderBy = "id"
)

// AllocationListRequestFilters contains the filters for loading multiple
// allocations.
type AllocationListRequestFilters struct {
	// StartDate is the first day of the window to report allocations for.
	//
	// The endpoint applies a default window when neither StartDate nor EndDate
	// is set: today through thirty days from today. It is not "everything", and
	// nothing in the response says the range was narrowed, so set both bounds
	// explicitly whenever the caller has a period in mind.
	StartDate *twapi.Date

	// EndDate is the last day of the window to report allocations for. See
	// StartDate for the default window.
	EndDate *twapi.Date

	// SearchTerm filters allocations by title.
	SearchTerm string

	// AssignedUserIDs filters allocations by the user whose time is committed.
	AssignedUserIDs []int64

	// AssignedUserTeamIDs filters allocations by the team of the assigned user.
	AssignedUserTeamIDs []int64

	// ProjectIDs filters allocations by project.
	ProjectIDs []int64

	// ProjectOwnerIDs filters allocations by the owner of their project.
	ProjectOwnerIDs []int64

	// ProjectCategoryIDs filters allocations by the category of their project.
	ProjectCategoryIDs []int64

	// ProjectCompanyIDs filters allocations by the company of their project.
	ProjectCompanyIDs []int64

	// ProjectTagIDs filters allocations by the tags of their project.
	ProjectTagIDs []int64

	// MatchAllProjectTags requires a project to carry every tag in
	// ProjectTagIDs rather than any of them.
	MatchAllProjectTags *bool

	// ProjectStatus filters allocations by the status of their project.
	ProjectStatus AllocationProjectStatus

	// ProjectHealths filters allocations by the health of their project: 0 for
	// not set, 1 for bad, 2 for ok and 3 for good.
	ProjectHealths []int64

	// OnlyStarredProjects restricts the results to starred projects.
	OnlyStarredProjects *bool

	// HideObservedProjects drops projects where the caller is only an observer.
	HideObservedProjects *bool

	// OnlyProjectsWithExplicitMembership restricts the results to projects the
	// caller is an explicit member of.
	OnlyProjectsWithExplicitMembership *bool

	// UpdatedAfter returns only allocations updated after this moment.
	UpdatedAfter *time.Time

	// DeletedAfter returns only allocations deleted after this moment. Pair it
	// with ShowDeleted, which is what makes deleted rows visible at all.
	DeletedAfter *time.Time

	// ShowDeleted includes soft-deleted allocations in the results. Without it
	// a deleted allocation cannot be found, which is what a restore needs.
	ShowDeleted *bool

	// TemplateWorkingHours skips working-hours processing and treats every day
	// as a standard working day.
	TemplateWorkingHours *bool

	// Include is the list of related entities to sideload into the response.
	Include []AllocationSideload

	// OrderBy is the field to sort the results by. Use the AllocationOrderBy
	// constants. The endpoint defaults to project.
	OrderBy AllocationOrderBy

	// OrderMode is the direction to sort the results in. See twapi.OrderMode for
	// the supported values. The endpoint defaults to ascending.
	OrderMode twapi.OrderMode

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of allocations to retrieve per page. Defaults to
	// 50.
	PageSize int64

	// CountMode selects whether the API computes the exact number of
	// allocations matching the filters, reported in Meta.Page.Count. Defaults
	// to twapi.ListCountModeDefault, which leaves the decision to the API.
	CountMode twapi.ListCountMode

	// Fields restricts the attributes returned for the allocation and each of
	// its sideloads. Each slot of AllocationListFields is a separate
	// `fields[entity]=…` selection; populated slots restrict the response, empty
	// slots return the API default. Use the generated AllocationField constants
	// to ensure values match real attributes.
	Fields AllocationListFields
}

func (a AllocationListRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	querySetDate(query, "startDate", a.StartDate)
	querySetDate(query, "endDate", a.EndDate)
	querySetString(query, "searchTerm", a.SearchTerm)
	querySetInt64s(query, "assignedUserIds", a.AssignedUserIDs)
	querySetInt64s(query, "assignedUserTeamIds", a.AssignedUserTeamIDs)
	querySetInt64s(query, "projectIds", a.ProjectIDs)
	querySetInt64s(query, "projectOwnerIds", a.ProjectOwnerIDs)
	querySetInt64s(query, "projectCategoryIds", a.ProjectCategoryIDs)
	querySetInt64s(query, "projectCompanyIds", a.ProjectCompanyIDs)
	querySetInt64s(query, "projectTagIds", a.ProjectTagIDs)
	querySetBool(query, "matchAllProjectTags", a.MatchAllProjectTags)
	querySetString(query, "projectStatus", a.ProjectStatus)
	querySetInt64s(query, "projectHealths", a.ProjectHealths)
	querySetBool(query, "onlyStarredProjects", a.OnlyStarredProjects)
	querySetBool(query, "hideObservedProjects", a.HideObservedProjects)
	querySetBool(query, "onlyProjectsWithExplicitMembership", a.OnlyProjectsWithExplicitMembership)
	querySetTimestamp(query, "updatedAfter", a.UpdatedAfter)
	querySetTimestamp(query, "deletedAfter", a.DeletedAfter)
	querySetBool(query, "showDeleted", a.ShowDeleted)
	querySetBool(query, "templateWorkingHours", a.TemplateWorkingHours)
	querySetStrings(query, "include", a.Include)
	querySetString(query, "orderBy", a.OrderBy)
	querySetString(query, "orderMode", a.OrderMode)
	querySetInt64(query, "page", a.Page)
	querySetInt64(query, "pageSize", a.PageSize)
	a.Fields.apply(query)
	a.CountMode.Apply(query)
	req.URL.RawQuery = query.Encode()
}

// AllocationListRequest represents the request for loading multiple
// allocations.
type AllocationListRequest struct {
	// Filters contains the filters for loading multiple allocations.
	Filters AllocationListRequestFilters
}

// NewAllocationListRequest creates a new AllocationListRequest with default
// values.
func NewAllocationListRequest() AllocationListRequest {
	return AllocationListRequest{
		Filters: AllocationListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the AllocationListRequest.
func (a AllocationListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/allocations.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	a.Filters.apply(req)

	return req, nil
}

// AllocationListResponse contains information by multiple allocations matching
// the request filters.
//
// sparsefields:list
type AllocationListResponse struct {
	request AllocationListRequest

	Meta        twapi.ListMeta `json:"meta"`
	Allocations []Allocation   `json:"allocations"`

	// Included contains the related objects requested through Include.
	//
	// The rate objects the financialDetails sideload also populates —
	// effectiveUserRate and costRate — are deliberately absent: the figures they
	// carry are already summarised inline on the allocation's FinancialDetails,
	// and modelling them would mean two new entities for the detail that is left.
	Included struct {
		// Projects contains the projects the allocations are committed to, keyed
		// by the string representation of the project ID.
		Projects map[string]Project `json:"projects,omitempty"`

		// Companies contains the companies owning those projects, keyed by the
		// string representation of the company ID.
		Companies map[string]Company `json:"companies,omitempty"`

		// Users contains the assigned users, keyed by the string representation of
		// the user ID. A user here may be a placeholder rather than a real person.
		Users map[string]User `json:"users,omitempty"`

		// JobRoles contains the job roles of the assigned users, keyed by the
		// string representation of the job role ID.
		//
		// The response spells this key jobRoles while the endpoint reads the
		// sparse fieldset from fields[jobroles], so the entity name is overridden
		// rather than inherited from the tag.
		//
		// sparsefields:key=jobroles
		JobRoles map[string]JobRole `json:"jobRoles,omitempty"`

		// Tasks contains the assigned users' tasks, keyed by the string
		// representation of the task ID.
		Tasks map[string]Task `json:"tasks,omitempty"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the AllocationListResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (a *AllocationListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list allocations")
	}
	if err := json.NewDecoder(resp.Body).Decode(a); err != nil {
		return fmt.Errorf("failed to decode list allocations response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (a *AllocationListResponse) SetRequest(req AllocationListRequest) {
	a.request = req
	a.Meta.ResolveCount(req.Filters.CountMode)
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (a *AllocationListResponse) Iterate() *AllocationListRequest {
	if !a.Meta.Page.HasMore {
		return nil
	}
	req := a.request
	req.Filters.Page++
	return &req
}

// AllocationList retrieves multiple allocations using the provided request and
// returns the response.
func AllocationList(
	ctx context.Context,
	engine *twapi.Engine,
	req AllocationListRequest,
) (*AllocationListResponse, error) {
	return twapi.Execute[AllocationListRequest, *AllocationListResponse](ctx, engine, req)
}
