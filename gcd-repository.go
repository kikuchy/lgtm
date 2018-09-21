// +build appengine

package main

import (
	"context"
	"errors"
	"github.com/labstack/echo"
	context2 "golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"math/rand"
	"time"
)

const KindImage string = "Image"
const KindServer string = "Server"

type gcdRepository struct {
	c context.Context
}

func (r *gcdRepository) LoadServers() ([]Server, error) {
	var servers []Server
	if _, err := datastore.NewQuery(KindServer).Order("-UpdateAt").GetAll(r.c, servers); err != nil {
		return nil, err
	}
	return servers, nil
}

func (r *gcdRepository) FindImageById(id int64) (*Image, error) {
	key := datastore.NewKey(r.c, KindImage, "", id, nil)
	var image Image
	if err := datastore.Get(r.c, key, &image); err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *gcdRepository) RandomImageId() (int64, error) {
	total, err := datastore.NewQuery(KindImage).Count(r.c)
	if err != nil {
		return -1, err
	}
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(total)
	t := datastore.NewQuery(KindImage).Offset(index).Limit(1).KeysOnly().Run(r.c)
	key, err := t.Next(nil)
	if err != nil {
		return -1, err
	}
	return key.IntID(), nil
}

func (r *gcdRepository) LoadImages(count uint, page uint) ([]Image, error) {
	query := datastore.
		NewQuery(KindImage).
		Limit(int(count)).
		Offset(int(count*(page-1))).
		Order("-CreatedAt").
		Filter("IsDeleted =", false)
	var images []Image
	if _, err := query.GetAll(r.c, &images); err != nil {
		return nil, err
	}
	return images, nil
}

func (r *gcdRepository) SaveImage(image *Image) error {
	return datastore.RunInTransaction(r.c, func(tc context2.Context) error {
		exists, err := datastore.NewQuery(KindImage).Filter("Url =", image.Url).Count(r.c)
		if err != nil {
			return err
		}
		if exists > 0 {
			// TODO: Error実装型にして重複時に専用のエラーメッセージ出せるようにしたい
			return errors.New("Duplicating URL: " + image.Url)
		}
		key := datastore.NewIncompleteKey(r.c, KindImage, nil)
		if _, err := datastore.Put(r.c, key, image); err != nil {
			return err
		}
		return nil
	}, nil)
}

func RepositoryGenerator(c echo.Context) Repository {
	return &gcdRepository{
		c: appengine.NewContext(c.Request()),
	}
}
