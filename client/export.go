package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	"go.icco.me/etu-backend/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// Export streams all of the caller's data to w as a JSON array of export items
// and returns the number of items written.
func (c *Config) Export(ctx context.Context, w io.Writer) (int, error) {
	userID, err := c.ensureUserID(ctx)
	if err != nil {
		return 0, err
	}
	g, err := c.getGRPCClients()
	if err != nil {
		return 0, err
	}
	stream, err := g.accountClient.ExportData(ctx, &proto.ExportDataRequest{UserId: userID})
	if err != nil {
		return 0, fmt.Errorf("export data: %w", err)
	}
	n, err := writeExport(w, stream.Recv)
	if err != nil {
		return n, fmt.Errorf("export data: %w", err)
	}
	return n, nil
}

// writeExport writes items from next to w as a JSON array until next returns io.EOF.
func writeExport(w io.Writer, next func() (*proto.ExportDataResponse, error)) (int, error) {
	bw := bufio.NewWriter(w)
	if _, err := bw.WriteString("["); err != nil {
		return 0, err
	}
	n := 0
	for {
		item, err := next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return n, err
		}
		data, err := protojson.Marshal(item)
		if err != nil {
			return n, err
		}
		sep := "\n"
		if n > 0 {
			sep = ",\n"
		}
		if _, err := bw.WriteString(sep); err != nil {
			return n, err
		}
		if _, err := bw.Write(data); err != nil {
			return n, err
		}
		n++
	}
	if _, err := bw.WriteString("\n]\n"); err != nil {
		return n, err
	}
	return n, bw.Flush()
}
