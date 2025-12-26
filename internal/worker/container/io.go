package container

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/moby/moby/client"
)

type IContainerIO interface {
	CopyTo()
	CopyFrom()
	DownloadIntoContainer()
	UploadFromContainer()
}

type ContainerIO struct {
	client *client.Client
}

func (c *ContainerIO) DownloadIntoContainer(ctx context.Context, containerID, url, path, filename string) error {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("DownloadIntoContainer: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("DownloadIntoContainer: %w", err)
	}

	if resp.ContentLength < 0 {
		return fmt.Errorf("DownloadIntoContainer: content length is less than 0")
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("DownloadIntoContainer: GET %s returned %d: %s", url, resp.StatusCode, string(b))
	}

	size := resp.ContentLength

	pr, pw := io.Pipe()
	tarDone := make(chan error, 1)
	go func() {
		defer resp.Body.Close()
		defer pw.Close()
		tw := tar.NewWriter(pw)
		defer tw.Close()

		hdr := &tar.Header{
			Name: filename,
			Mode: 0600,
			Size: size,
		}

		if err := tw.WriteHeader(hdr); err != nil {
			tarDone <- err
			pw.CloseWithError(err)
			return
		}

		if _, err := io.Copy(tw, resp.Body); err != nil {
			tarDone <- err
			pw.CloseWithError(err)
			return
		}
	}()

	_, copyErr := c.client.CopyToContainer(ctx, containerID, client.CopyToContainerOptions{
		Content:         pr,
		DestinationPath: path,
	})

	if copyErr != nil {
		return copyErr
	}

	return nil
}

func (c *ContainerIO) UploadFromContainer(ctx context.Context, containerID, url, path, filename string) error {
	res, err := c.client.CopyFromContainer(ctx, containerID, client.CopyFromContainerOptions{
		SourcePath: path,
	})

	if err != nil {
		return err
	}

	defer res.Content.Close()

	tr := tar.NewReader(res.Content)
	var hdr *tar.Header

	for {
		h, err := tr.Next()

		if err == io.EOF {
			return fmt.Errorf("file %s not found in tar", filename)
		}

		if err != nil {
			return err
		}

		if h.Typeflag == tar.TypeReg && h.Name == filename {
			hdr = h
			break
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, tr)
	if err != nil {
		return err
	}

	req.ContentLength = hdr.Size

	client := http.Client{
		Timeout: 10 * time.Minute,
	}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil

}
