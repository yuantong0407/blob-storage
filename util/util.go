package util

import (
	"crypto/rand"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	ds "github.com/enirinth/blob-storage/clusterds"
	"io"
	"strconv"
)

const (
	separator string = "-----"
)

// UUid generator
// newUUID generates a random UUID according to RFC 4122
// Credit to https://play.golang.org/p/4FkNSiUDMg
func NewUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

// Print storage table for a certain DC
func PrintStorage(storageTable *map[string]*ds.Partition) {
	for _, v := range *storageTable {
		fmt.Println("Partition with ID: " + (*v).PartitionID + " starts" + separator)
		fmt.Println("Partition size: " + strconv.FormatFloat((*v).PartitionSize, 'f', 6, 64) + " ; partition createtimestamp: " + strconv.FormatInt((*v).CreateTimestamp, 10))
		for _, blob := range (*v).BlobList {
			fmt.Println(blob)
		}
		fmt.Println("Partition ends" + separator)
	}
}

// Find if a certain DC stores a certain partition
func FindDC(dcID string, pState *ds.PartitionState) bool {
	for _, id := range (*pState).DCList {
		if id == dcID {
			return true
		}
	}
	return false
}

// Find blob (using its ID) in a partition
// return True if found
func FindBlob(blobID string, partition *ds.Partition) bool {
	for _, blob := range partition.BlobList {
		if blobID == blob.BlobID {
			return true
		}
	}
	return false
}

// Merge one partition into another, so that contains (in set semantics) all the blobs
// Happens during inter-DC synchronization
func MergePartition(p1 *ds.Partition, p2 *ds.Partition) {
	if p1.PartitionID != p2.PartitionID {
		err := errors.New("Cannot merge two different partitions (with different IDs)")
		log.Fatal(err)
	}
	for _, blob := range p2.BlobList {
		if !FindBlob(blob.BlobID, p1) {
			p1.AppendBlob(blob)
		}
	}
}

func FindPartition(partitionID string, m *map[string]*ds.PartitionState) bool {
	_, ok := (*m)[partitionID]
	return ok
}
