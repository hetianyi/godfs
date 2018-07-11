package header

// Definition some data headers for data transfer.
// [operation][meta][body...]
type UploadHeader struct {
    operation [8]byte   //operations such as upload, sync, communication
    metaLen [10]
    meta []
    bodyLen []
    body []
}