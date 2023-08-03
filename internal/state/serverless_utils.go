package state

import (
	"context"
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetServerlessOrServed(ctx context.Context, req ctrl.Request, c client.Client) (*v1alpha1.Serverless, error) {
	instance := &v1alpha1.Serverless{}
	err := c.Get(ctx, req.NamespacedName, instance)
	if err == nil {
		return instance, nil
	}
	if !k8serrors.IsNotFound(err) {
		return nil, errors.Wrap(err, "while fetching serverless instance")
	}

	instance, err = GetServedServerless(ctx, c)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching served serverless instance")
	}
	return instance, nil
}

func GetServedServerless(ctx context.Context, c client.Client) (*v1alpha1.Serverless, error) {
	var serverlessList v1alpha1.ServerlessList

	err := c.List(ctx, &serverlessList)

	if err != nil {
		return nil, err
	}

	for _, item := range serverlessList.Items {
		if !item.IsServedEmpty() && item.Status.Served == v1alpha1.ServedTrue {
			return &item, nil
		}
	}

	return nil, nil
}
