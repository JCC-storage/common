package io2

type errorWriter struct {
	err error
}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, w.err
}

func (w *errorWriter) Close() error {
	return nil
}

func ErrorWriter(err error) *errorWriter {
	return &errorWriter{err: err}
}

type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

func (r *errorReader) Close() error {
	return nil
}

func ErrorReader(err error) *errorReader {
	return &errorReader{err: err}
}
