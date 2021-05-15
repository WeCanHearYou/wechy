package actions

import (
	"context"

	"github.com/getfider/fider/app/models/entities"
	"github.com/getfider/fider/app/models/enum"
	"github.com/getfider/fider/app/models/query"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/gosimple/slug"

	"github.com/getfider/fider/app"
	"github.com/getfider/fider/app/models"
	"github.com/getfider/fider/app/pkg/errors"
	"github.com/getfider/fider/app/pkg/validate"
)

// CreateNewPost is used to create a new post
type CreateNewPost struct {
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Attachments []*models.ImageUpload `json:"attachments"`
}

// IsAuthorized returns true if current user is authorized to perform this action
func (action *CreateNewPost) IsAuthorized(ctx context.Context, user *entities.User) bool {
	return user != nil
}

// Validate if current model is valid
func (action *CreateNewPost) Validate(ctx context.Context, user *entities.User) *validate.Result {
	result := validate.Success()

	if action.Title == "" {
		result.AddFieldFailure("title", "Title is required.")
	} else if len(action.Title) < 10 {
		result.AddFieldFailure("title", "Title needs to be more descriptive.")
	} else if len(action.Title) > 100 {
		result.AddFieldFailure("title", "Title must have less than 100 characters.")
	} else {
		err := bus.Dispatch(ctx, &query.GetPostBySlug{Slug: slug.Make(action.Title)})
		if err != nil && errors.Cause(err) != app.ErrNotFound {
			return validate.Error(err)
		} else if err == nil {
			result.AddFieldFailure("title", "This has already been posted before.")
		}
	}

	messages, err := validate.MultiImageUpload(nil, action.Attachments, validate.MultiImageUploadOpts{
		MaxUploads:   3,
		MaxKilobytes: 5120,
		ExactRatio:   false,
	})
	if err != nil {
		return validate.Error(err)
	}
	result.AddFieldFailure("attachments", messages...)

	return result
}

// UpdatePost is used to edit an existing new post
type UpdatePost struct {
	Number      int                   `route:"number"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Attachments []*models.ImageUpload `json:"attachments"`

	Post *models.Post
}

// IsAuthorized returns true if current user is authorized to perform this action
func (action *UpdatePost) IsAuthorized(ctx context.Context, user *entities.User) bool {
	return user != nil && user.IsCollaborator()
}

// Validate if current model is valid
func (action *UpdatePost) Validate(ctx context.Context, user *entities.User) *validate.Result {
	result := validate.Success()

	if action.Title == "" {
		result.AddFieldFailure("title", "Title is required.")
	} else if len(action.Title) < 10 {
		result.AddFieldFailure("title", "Title needs to be more descriptive.")
	}

	if len(action.Title) > 100 {
		result.AddFieldFailure("title", "Title must have less than 100 characters.")
	}

	postByNumber := &query.GetPostByNumber{Number: action.Number}
	if err := bus.Dispatch(ctx, postByNumber); err != nil {
		return validate.Error(err)
	}

	action.Post = postByNumber.Result

	postBySlug := &query.GetPostBySlug{Slug: slug.Make(action.Title)}
	err := bus.Dispatch(ctx, postBySlug)
	if err != nil && errors.Cause(err) != app.ErrNotFound {
		return validate.Error(err)
	} else if err == nil && postBySlug.Result.ID != action.Post.ID {
		result.AddFieldFailure("title", "This has already been posted before.")
	}

	if len(action.Attachments) > 0 {
		getAttachments := &query.GetAttachments{Post: action.Post}
		err = bus.Dispatch(ctx, getAttachments)
		if err != nil {
			return validate.Error(err)
		}

		messages, err := validate.MultiImageUpload(getAttachments.Result, action.Attachments, validate.MultiImageUploadOpts{
			MaxUploads:   3,
			MaxKilobytes: 5120,
			ExactRatio:   false,
		})
		if err != nil {
			return validate.Error(err)
		}
		result.AddFieldFailure("attachments", messages...)
	}

	return result
}

// AddNewComment represents a new comment to be added
type AddNewComment struct {
	Number      int                   `route:"number"`
	Content     string                `json:"content"`
	Attachments []*models.ImageUpload `json:"attachments"`
}

// IsAuthorized returns true if current user is authorized to perform this action
func (action *AddNewComment) IsAuthorized(ctx context.Context, user *entities.User) bool {
	return user != nil
}

// Validate if current model is valid
func (action *AddNewComment) Validate(ctx context.Context, user *entities.User) *validate.Result {
	result := validate.Success()

	if action.Content == "" {
		result.AddFieldFailure("content", "Comment is required.")
	}

	messages, err := validate.MultiImageUpload(nil, action.Attachments, validate.MultiImageUploadOpts{
		MaxUploads:   2,
		MaxKilobytes: 5120,
		ExactRatio:   false,
	})
	if err != nil {
		return validate.Error(err)
	}
	result.AddFieldFailure("attachments", messages...)

	return result
}

// SetResponse represents the action to update an post response
type SetResponse struct {
	Number         int             `route:"number"`
	Status         enum.PostStatus `json:"status"`
	Text           string          `json:"text"`
	OriginalNumber int             `json:"originalNumber"`

	Original *models.Post
}

// IsAuthorized returns true if current user is authorized to perform this action
func (action *SetResponse) IsAuthorized(ctx context.Context, user *entities.User) bool {
	return user != nil && user.IsCollaborator()
}

// Validate if current model is valid
func (action *SetResponse) Validate(ctx context.Context, user *entities.User) *validate.Result {
	result := validate.Success()

	if action.Status < enum.PostOpen || action.Status > enum.PostDuplicate {
		result.AddFieldFailure("status", "Status is invalid.")
	}

	if action.Status == enum.PostDuplicate {
		if action.OriginalNumber == action.Number {
			result.AddFieldFailure("originalNumber", "Cannot be a duplicate of itself")
		}

		getOriginaPost := &query.GetPostByNumber{Number: action.OriginalNumber}
		err := bus.Dispatch(ctx, getOriginaPost)
		if err != nil {
			if errors.Cause(err) == app.ErrNotFound {
				result.AddFieldFailure("originalNumber", "Original post not found")
			} else {
				return validate.Error(err)
			}
		}

		if getOriginaPost.Result != nil {
			action.Original = getOriginaPost.Result
		}
	}

	return result
}

// DeletePost represents the action of an administrator deleting an existing Post
type DeletePost struct {
	Number int    `route:"number"`
	Text   string `json:"text"`

	Post *models.Post
}

// IsAuthorized returns true if current user is authorized to perform this action
func (action *DeletePost) IsAuthorized(ctx context.Context, user *entities.User) bool {
	return user != nil && user.IsAdministrator()
}

// Validate if current model is valid
func (action *DeletePost) Validate(ctx context.Context, user *entities.User) *validate.Result {
	getPost := &query.GetPostByNumber{Number: action.Number}
	if err := bus.Dispatch(ctx, getPost); err != nil {
		return validate.Error(err)
	}

	action.Post = getPost.Result

	isReferencedQuery := &query.PostIsReferenced{PostID: action.Post.ID}
	if err := bus.Dispatch(ctx, isReferencedQuery); err != nil {
		return validate.Error(err)
	}

	if isReferencedQuery.Result {
		return validate.Failed("This post cannot be deleted because it's being referenced by a duplicated post.")
	}

	return validate.Success()
}

// EditComment represents the action to update an existing comment
type EditComment struct {
	PostNumber  int                   `route:"number"`
	ID          int                   `route:"id"`
	Content     string                `json:"content"`
	Attachments []*models.ImageUpload `json:"attachments"`

	Post    *models.Post
	Comment *models.Comment
}

// IsAuthorized returns true if current user is authorized to perform this action
func (action *EditComment) IsAuthorized(ctx context.Context, user *entities.User) bool {
	postByNumber := &query.GetPostByNumber{Number: action.PostNumber}
	commentByID := &query.GetCommentByID{CommentID: action.ID}
	if err := bus.Dispatch(ctx, postByNumber, commentByID); err != nil {
		return false
	}

	action.Post = postByNumber.Result
	action.Comment = commentByID.Result
	return user.ID == action.Comment.User.ID || user.IsCollaborator()
}

// Validate if current model is valid
func (action *EditComment) Validate(ctx context.Context, user *entities.User) *validate.Result {
	result := validate.Success()

	if action.Content == "" {
		result.AddFieldFailure("content", "Comment is required.")
	}

	if len(action.Attachments) > 0 {
		getAttachments := &query.GetAttachments{Post: action.Post, Comment: action.Comment}
		err := bus.Dispatch(ctx, getAttachments)
		if err != nil {
			return validate.Error(err)
		}

		messages, err := validate.MultiImageUpload(getAttachments.Result, action.Attachments, validate.MultiImageUploadOpts{
			MaxUploads:   2,
			MaxKilobytes: 5120,
			ExactRatio:   false,
		})
		if err != nil {
			return validate.Error(err)
		}
		result.AddFieldFailure("attachments", messages...)
	}

	return result
}

// DeleteComment represents the action of deleting an existing comment
type DeleteComment struct {
	PostNumber int `route:"number"`
	CommentID  int `route:"id"`
}

// IsAuthorized returns true if current user is authorized to perform this action
func (action *DeleteComment) IsAuthorized(ctx context.Context, user *entities.User) bool {
	commentByID := &query.GetCommentByID{CommentID: action.CommentID}
	if err := bus.Dispatch(ctx, commentByID); err != nil {
		return false
	}

	return user.ID == commentByID.Result.User.ID || user.IsCollaborator()
}

// Validate if current model is valid
func (action *DeleteComment) Validate(ctx context.Context, user *entities.User) *validate.Result {
	return validate.Success()
}
