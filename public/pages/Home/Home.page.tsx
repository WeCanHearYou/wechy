import "./Home.page.scss"
import NoDataIllustration from "@fider/assets/images/undraw-no-data.svg"

import React, { useState } from "react"
import { Post, Tag, PostStatus } from "@fider/models"
import { Markdown, Hint, PoweredByFider, Icon } from "@fider/components"
import { SimilarPosts } from "./components/SimilarPosts"
import { PostInput } from "./components/PostInput"
import { PostsContainer } from "./components/PostsContainer"
import { useFider } from "@fider/hooks"
import { VStack } from "@fider/components/layout"
import { useTranslation } from "react-i18next"

export interface HomePageProps {
  posts: Post[]
  tags: Tag[]
  countPerStatus: { [key: string]: number }
  t: (key: string) => string
}

export interface HomePageState {
  title: string
}

const Lonely = () => {
  const fider = useFider()
  const { t } = useTranslation()

  return (
    <div className="text-center">
      <Hint permanentCloseKey="at-least-3-posts" condition={fider.session.isAuthenticated && fider.session.user.isAdministrator}>
        <p>
          <span
            dangerouslySetInnerHTML={{
              __html: t(
                "It&apos;s recommended that you create <strong>at least 3</strong> suggestions here before sharing this site. The initial content is important to start engaging your audience."
              ),
            }}
          />
        </p>
      </Hint>
      <Icon sprite={NoDataIllustration} height="120" className="mt-6 mb-2" />
      <p className="text-muted">{t("No posts have been created yet.")}</p>
    </div>
  )
}

const defaultWelcomeMessage = () => {
  const { t } = useTranslation()
  return t(`We'd love to hear what you're thinking about. \n\nWhat can we do better? This is the place for you to vote, discuss and share ideas.`)
}

const HomePage = (props: HomePageProps) => {
  const fider = useFider()
  const [title, setTitle] = useState("")
  const { t } = useTranslation()

  const isLonely = () => {
    const len = Object.keys(props.countPerStatus).length
    if (len === 0) {
      return true
    }

    if (len === 1 && PostStatus.Deleted.value in props.countPerStatus) {
      return true
    }

    return false
  }

  return (
    <div id="p-home" className="page container">
      <div className="p-home__welcome-col">
        <VStack spacing={2}>
          <Markdown text={fider.session.tenant.welcomeMessage || defaultWelcomeMessage()} style="full" />
          <PostInput placeholder={fider.session.tenant.invitation || t("Enter your suggestion here...")} onTitleChanged={setTitle} />
          <PoweredByFider className="sm:hidden md:hidden lg:block" />
        </VStack>
      </div>
      <div className="p-home__posts-col">
        {isLonely() ? (
          <Lonely />
        ) : title ? (
          <SimilarPosts title={title} tags={props.tags} />
        ) : (
          <PostsContainer posts={props.posts} tags={props.tags} countPerStatus={props.countPerStatus} />
        )}
        <PoweredByFider className="lg:hidden xl:hidden mt-8" />
      </div>
    </div>
  )
}

export default HomePage
