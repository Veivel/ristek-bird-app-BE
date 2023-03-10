package postcontroller

import (
	"bird-app/lib"
	"bird-app/models"
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GuestIndex(c *gin.Context) {

	c.JSON(200, gin.H{
		"message": "Not implemented yet",
	})
}

// Show user's "home timeline", including close friends.
func Index(c *gin.Context) {
	const POSTS_PER_PAGE = 10

	placeholderPost := models.Post{}
	posts := []models.Post{}
	pageStr, _ := strconv.Atoi(c.Query("page"))
	newNum := POSTS_PER_PAGE * (int64(math.Max(1.0, float64(pageStr))) - 1)

	cursor, err := lib.DB.Collection("posts").Find(context.Background(), bson.D{}, &options.FindOptions{
		Skip: &newNum,
		Sort: bson.D{{Key: "_id", Value: -1}},
	})
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"message": err.Error(),
		})
		return
	} else {
		for i := 0; i < POSTS_PER_PAGE && cursor.TryNext(context.Background()); i++ {
			posts = append(posts, placeholderPost)
			cursor.Decode(&posts[i])
			if posts[i].Author == "" {
				break
			}
		}
		c.JSON(200, gin.H{
			"data": posts,
		})
	}
}

func Show(c *gin.Context) {
	var post models.Post
	uuid := c.Param("postUuid")

	result := lib.DB.Collection("posts").FindOne(
		context.Background(),
		bson.D{{Key: "uuid", Value: uuid}},
	)

	if result.Err() != nil {
		c.AbortWithStatusJSON(404, gin.H{
			"message": "Could not find post with specified UUID.",
		})
		return
	}

	result.Decode(&post)
	c.JSON(200, gin.H{
		"data": post,
	})
}

// Submit new post
func Create(c *gin.Context) {
	// var post models.Post
	var body models.Post
	username, _ := c.Get("username")
	avatar, _ := c.Get("avatar")

	err := c.BindJSON(&body) // text, is_close_friends
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	post := models.Post{
		Uuid:           uuid.New().String(),
		Text:           body.Text,
		IsCloseFriends: body.IsCloseFriends,
		Author:         username.(string),
		AuthorAvatar:   avatar.(string),
		CreatedAt:      time.Now(),
		Comments:       []models.Comment{},
		Likes:          0,
	}

	lib.DB.Collection("posts").InsertOne(context.Background(), post)

	c.JSON(201, gin.H{
		"message": "Successfully created post.",
		"data":    post,
	})
}

// Edit post content
func Edit(c *gin.Context) {
	posts := lib.DB.Collection("posts")
	var body models.Post
	var post models.Post
	username, _ := c.Get("username")

	c.BindJSON(&body)

	criteria := bson.D{{Key: "uuid", Value: c.Param("postUuid")}}
	result := posts.FindOne(
		context.Background(),
		criteria,
	)

	if result.Err() != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"message": "Could not find post with specified UUID.",
		})
		return
	}

	result.Decode(&post)

	if username.(string) != post.Author {
		c.AbortWithStatusJSON(401, gin.H{
			"message": "You do not have access to edit this post.",
			"data": gin.H{
				"author": post.Author,
				"you":    username.(string),
			},
		})
		return
	}

	fmt.Println(body)
	if body.Text != "" {
		post.Text = body.Text
	}

	fmt.Println(post)
	posts.FindOneAndReplace(
		context.Background(),
		criteria,
		post,
	)

	c.JSON(200, gin.H{
		"message": "Post successfully edited",
		"data":    post,
	})
}

// Delete post
func Delete(c *gin.Context) {
	posts := lib.DB.Collection("posts")
	var post models.Post
	username, _ := c.Get("username")

	criteria := bson.D{{Key: "uuid", Value: c.Param("postUuid")}}
	result := posts.FindOne(
		context.Background(),
		criteria,
	)

	if result.Err() != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"message": "Could not find post with specified UUID.",
		})
		return
	}

	result.Decode(&post)

	if username.(string) != post.Author {
		c.AbortWithStatusJSON(401, gin.H{
			"message": "You do not have access to edit this post.",
			"data": gin.H{
				"author": post.Author,
				"you":    username.(string),
			},
		})
		return
	}

	posts.FindOneAndDelete(
		context.Background(),
		criteria,
	)

	c.JSON(200, gin.H{
		"message": "Post successfully deleted.",
	})
}
