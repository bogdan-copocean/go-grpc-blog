package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	protos "github.com/bogdan-user/go-grpc-react-blog/proto"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var Collection *mongo.Collection

type BlogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

type server struct {
	protos.UnimplementedBlogServiceServer
}

func NewServer() *server {
	return &server{}
}

type grpcMultiplexer struct {
	*grpcweb.WrappedGrpcServer
}

// Handler is used to route requests to either grpc or to regular http
func (m *grpcMultiplexer) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.IsGrpcWebRequest(r) {
			m.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (*server) CreateBlog(ctx context.Context, req *protos.CreateBlogRequest) (*protos.CreateBlogResponse, error) {
	fmt.Println("CreateBlog invoked")

	blog := req.GetBlog()

	data := BlogItem{
		AuthorID: blog.GetAuthorId(),
		Content:  blog.GetContent(),
		Title:    blog.GetTitle(),
	}

	res, err := Collection.InsertOne(ctx, data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal error: %v\n", err))
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot convert to OID: %v\n", err))
	}

	return &protos.CreateBlogResponse{
		Blog: &protos.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Content:  blog.GetContent(),
			Title:    blog.GetTitle(),
		},
	}, nil
}

func (*server) ReadBlog(ctx context.Context, req *protos.ReadBlogRequest) (*protos.ReadBlogResponse, error) {
	fmt.Println("ReadBlog invoked")

	blogId := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogId)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse id: %v.\n", err))
	}

	// create an empty struct
	data := &BlogItem{}
	err = Collection.FindOne(context.TODO(), bson.D{{"_id", oid}}).Decode(data)

	if err == mongo.ErrNoDocuments {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Blog with id not found: %v\n", err))
	}

	return &protos.ReadBlogResponse{
		Blog: &protos.Blog{
			Id:       data.ID.Hex(),
			AuthorId: data.AuthorID,
			Content:  data.Content,
			Title:    data.Title,
		},
	}, nil

}

func (*server) UpdateBlog(ctx context.Context, req *protos.UpdateBlogRequest) (*protos.UpdateBlogResponse, error) {
	fmt.Println("UpdateBlog invoked")
	blog := req.GetBlog()

	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse id: %v.\n", err))
	}

	data := BlogItem{}

	err = Collection.FindOne(context.TODO(), bson.D{{"_id", oid}}).Decode(data)

	if err == mongo.ErrNoDocuments {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Blog with id not found: %v\n", err))
	}

	// updating...
	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()
	data.ID = oid

	filter := bson.D{{"_id", oid}}

	_, err = Collection.ReplaceOne(context.TODO(), filter, data)

	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot update object in MongoDB: %v\n", err))
	}

	return &protos.UpdateBlogResponse{
		Blog: &protos.Blog{
			Id:       data.ID.Hex(),
			AuthorId: data.AuthorID,
			Content:  data.Content,
			Title:    data.Title,
		},
	}, nil

}

func (*server) DeleteBlog(ctx context.Context, req *protos.DeleteBlogRequest) (*protos.DeleteBlogResponse, error) {
	fmt.Println("DeleteBlog invoked")

	blogId := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse id: %v.\n", err))
	}

	data := BlogItem{}
	err = Collection.FindOne(context.TODO(), bson.D{{"_id", oid}}).Decode(data)

	if err == mongo.ErrNoDocuments {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Blog with id not found: %v\n", err))
	}

	filter := bson.D{{"_id", oid}}

	_, err = Collection.DeleteOne(context.TODO(), filter)

	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot delete from MongoDB: %v\n", err))
	}

	return &protos.DeleteBlogResponse{
		BlogId: oid.Hex(),
	}, nil

}

func (*server) ListBlog(req *protos.ListBlogRequest, stream protos.BlogService_ListBlogServer) error {
	fmt.Println("ListAllBlog invoked")

	filter := bson.D{}

	cur, err := Collection.Find(context.TODO(), filter)
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Uknown internal err: %v\n", err))
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		data := &BlogItem{}
		err := cur.Decode(data)
		if err != nil {
			return status.Errorf(codes.Internal, fmt.Sprintf("Err while decoding data: %v\n", err))
		}

		stream.Send(&protos.ListBlogResponse{
			Blog: &protos.Blog{
				Id:       data.ID.Hex(),
				AuthorId: data.AuthorID,
				Content:  data.Content,
				Title:    data.Title,
			},
		})
		time.Sleep(1 * time.Second)
	}
	if err := cur.Err(); err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Error final step: %v\n", err))
	}

	return nil
}
