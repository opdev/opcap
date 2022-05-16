/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"google.golang.org/grpc"

	"github.com/operator-framework/operator-registry/pkg/api"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"context"
	"io"
	"time"

	"github.com/operator-framework/operator-registry/pkg/api/grpc_health_v1"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("check called")
		c, err := NewClient("172.30.180.138:50051")
		if err != nil {
			fmt.Println(err)
		}
		bundles, err := c.ListBundles(context.Background())
		if err != nil {
			fmt.Println(err)
		}
		for {
			b := bundles.Next()
			if b == nil {
				break
			}
			fmt.Println(b.CsvName)
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type Client struct {
	Registry api.RegistryClient
	Health   grpc_health_v1.HealthClient
	Conn     *grpc.ClientConn
}

type Interface interface {
	GetBundle(ctx context.Context, packageName, channelName, csvName string) (*api.Bundle, error)
	GetBundleInPackageChannel(ctx context.Context, packageName, channelName string) (*api.Bundle, error)
	GetReplacementBundleInPackageChannel(ctx context.Context, currentName, packageName, channelName string) (*api.Bundle, error)
	GetBundleThatProvides(ctx context.Context, group, version, kind string) (*api.Bundle, error)
	ListBundles(ctx context.Context) (*BundleIterator, error)
	GetPackage(ctx context.Context, packageName string) (*api.Package, error)
	HealthCheck(ctx context.Context, reconnectTimeout time.Duration) (bool, error)
	Close() error
}

// var _ Interface = &Client{}

type BundleStream interface {
	Recv() (*api.Bundle, error)
}

type BundleIterator struct {
	stream BundleStream
	error  error
}

func NewBundleIterator(stream BundleStream) *BundleIterator {
	return &BundleIterator{stream: stream}
}

func (it *BundleIterator) Next() *api.Bundle {
	if it.error != nil {
		return nil
	}
	next, err := it.stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		it.error = err
	}
	return next
}

func (it *BundleIterator) Error() error {
	return it.error
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return NewClientFromConn(conn), nil
}

func NewClientFromConn(conn *grpc.ClientConn) *Client {
	return &Client{
		Registry: api.NewRegistryClient(conn),
		Health:   grpc_health_v1.NewHealthClient(conn),
		Conn:     conn,
	}
}

func (c *Client) ListBundles(ctx context.Context) (*BundleIterator, error) {
	stream, err := c.Registry.ListBundles(ctx, &api.ListBundlesRequest{})
	if err != nil {
		return nil, err
	}
	return NewBundleIterator(stream), nil
}

// NewClient
func GetK8sClient() *kubernetes.Clientset {
	// create k8s client
	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		_ = fmt.Errorf("unable to build config from flags: %v", err)
	}
	clientset, _ := kubernetes.NewForConfig(cfg)

	return clientset
}

func GetTest(c *kubernetes.Clientset) {

	// temporary api commnication test
	ctx := context.Background()
	nodes, _ := c.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	for _, n := range nodes.Items {
		fmt.Print(n.ObjectMeta.Name)
	}

}
