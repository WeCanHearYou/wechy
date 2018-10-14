import * as React from "react";
import { Comment, CurrentUser, Post } from "@fider/models";
import { Gravatar, UserName, Moment, Form, TextArea, Button, MultiLineText, DropDown, DropDownItem } from "@fider/components";
import { formatDate, Failure, actions, Fider } from "@fider/services";

interface ShowCommentProps {
  post: Post;
  comment: Comment;
}

interface ShowCommentState {
  comment: Comment;
  isEditing: boolean;
  newContent: string;
  error?: Failure;
}

export class ShowComment extends React.Component<ShowCommentProps, ShowCommentState> {
  constructor(props: ShowCommentProps) {
    super(props);
    this.state = {
      comment: props.comment,
      isEditing: false,
      newContent: ""
    };
  }

  private canEditComment(comment: Comment): boolean {
    if (Fider.session.isAuthenticated) {
      return Fider.session.user.isCollaborator || comment.user.id === Fider.session.user.id;
    }
    return false;
  }

  private cancelEdit = async () => {
    this.setState({
      isEditing: false,
      newContent: "",
      error: undefined
    });
  };

  private saveEdit = async () => {
    const response = await actions.updateComment(this.props.post.number, this.state.comment.id, this.state.newContent);
    if (response.ok) {
      this.state.comment.content = this.state.newContent;
      this.state.comment.editedAt = new Date().toISOString();
      this.state.comment.editedBy = Fider.session.user;
      this.setState({
        comment: this.state.comment
      });
      this.cancelEdit();
    } else {
      this.setState({ error: response.error });
    }
  };

  private setNewContent = (newContent: string) => {
    this.setState({ newContent });
  };

  private renderText = () => {
    return <i className="ellipsis horizontal icon" />;
  };

  private onActionSelected = (item: DropDownItem) => {
    if (item.value === "edit") {
      this.setState({ isEditing: true, newContent: this.state.comment.content, error: undefined });
    }
  };

  public render() {
    const c = this.state.comment;

    const editedMetadata = !!c.editedAt &&
      !!c.editedBy && (
        <div className="c-comment-metadata">
          <span title={`This comment has been edited by ${c.editedBy!.name} on ${formatDate(c.editedAt)}`}>edited</span>
        </div>
      );

    return (
      <div className="c-comment">
        <div className="c-comment-header">
          <Gravatar user={c.user} />
          <div className="c-comment-title">
            <UserName user={c.user} />
            <div className="c-comment-metadata">
              <Moment date={c.createdAt} />
            </div>
            {editedMetadata}
          </div>
          {!this.state.isEditing &&
            this.canEditComment(c) && (
              <DropDown
                className="l-more-actions"
                direction="left"
                items={[{ label: "Edit", value: "edit" }]}
                onChange={this.onActionSelected}
                renderText={this.renderText}
              />
            )}
        </div>
        <div className="c-comment-text">
          {this.state.isEditing ? (
            <Form error={this.state.error}>
              <TextArea
                field="content"
                minRows={1}
                value={this.state.newContent}
                placeholder={c.content}
                onChange={this.setNewContent}
              />
              <Button size="tiny" onClick={this.saveEdit} color="positive">
                Save
              </Button>
              <Button color="cancel" size="tiny" onClick={this.cancelEdit}>
                Cancel
              </Button>
            </Form>
          ) : (
            <MultiLineText text={c.content} style="simple" />
          )}
        </div>
      </div>
    );
  }
}
