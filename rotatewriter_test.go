package rotatewriter

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

func cleanupLogs() {
	files, err := ioutil.ReadDir("logs")
	if err != nil {
		return
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if (f.Name() == ".gitignore") || (f.Name() == "README") {
			continue
		}
		os.Remove(path.Join("logs", f.Name()))
	}
	os.Remove("logs/test.log")
}

func compareRws(rw1, rw2 *RotateWriter) bool {
	if (rw1 == nil) && (rw2 == nil) {
		return true
	}
	if (rw1 == nil) || (rw2 == nil) {
		return false
	}
	return (rw1.Filename == rw2.Filename) && (rw1.NumFiles == rw2.NumFiles) && (rw1.dirpath == rw2.dirpath)
}

func TestNewRotateWriter(t *testing.T) {
	type args struct {
		fname    string
		numfiles int
	}
	tests := []struct {
		name    string
		args    args
		wantRw  *RotateWriter
		wantErr bool
		prepare func()
		cleanUp func()
	}{
		{
			name:    "Empty filaname case",
			args:    args{fname: "", numfiles: 0},
			wantRw:  nil,
			wantErr: true,
		},
		{
			name:    "Not existing path to log file",
			args:    args{"/notexisting/log/path/of/file", 0},
			wantRw:  nil,
			wantErr: true,
		},
		{
			name:    "Log file path exists all ok, but negative numfiles",
			args:    args{fname: "logs/test.log", numfiles: -1},
			wantRw:  nil,
			wantErr: true,
		},
		{
			name: "Log file path exists, zero numfiles",
			args: args{fname: "logs/test.log", numfiles: 0},
			wantRw: &RotateWriter{
				Filename: "logs/test.log",
				NumFiles: 0,
				dirpath:  "logs",
			},
			wantErr: false,
			cleanUp: cleanupLogs,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRw, err := NewRotateWriter(tt.args.fname, tt.args.numfiles)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRotateWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareRws(gotRw, tt.wantRw) {
				t.Errorf("NewRotateWriter() = %v, want %v", gotRw, tt.wantRw)
			}
			if nil != tt.cleanUp {
				tt.cleanUp()
			}
		})
	}
}

func TestRotateWriter_initDirPath(t *testing.T) {
	type fields struct {
		Filename string
		NumFiles int
		dirpath  string
	}
	defStartUpfunc := func(fields fields) *RotateWriter {
		return &RotateWriter{
			Filename: fields.Filename,
			NumFiles: fields.NumFiles,
		}
	}
	tests := []struct {
		name        string
		fields      fields
		wantErr     bool
		wantRw      *RotateWriter
		startUpFunc func(fields fields) *RotateWriter
		cleanUp     func()
	}{
		{
			name:        "Empty path",
			fields:      fields{Filename: "", NumFiles: 0, dirpath: ""},
			startUpFunc: nil,
			wantErr:     true,
			wantRw:      nil,
			cleanUp:     cleanupLogs,
		},
		{
			name:        "Normal path",
			fields:      fields{Filename: "logs/test.log", NumFiles: 0, dirpath: ""},
			startUpFunc: nil,
			wantErr:     false,
			wantRw:      &RotateWriter{Filename: "logs/test.log", NumFiles: 0, dirpath: "logs"},
			cleanUp:     cleanupLogs,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rw *RotateWriter
			rw = nil
			if tt.startUpFunc == nil {
				rw = defStartUpfunc(tt.fields)
			} else {
				rw = tt.startUpFunc(tt.fields)
			}
			err := rw.initDirPath()
			if (err != nil) != tt.wantErr {
				t.Errorf("RotateWriter.initDirPath() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err == nil) && !tt.wantErr {
				if !compareRws(rw, tt.wantRw) {
					t.Errorf("RotateWriter.initDirPath() = %v, want %v", rw, tt.wantRw)
				}
			}
		})
	}
}

func TestRotateWriter_openWriteFile(t *testing.T) {
	type fields struct {
		Filename string
		NumFiles int
		dirpath  string
	}
	defStartUpfunc := func(fields fields) *RotateWriter {
		return &RotateWriter{
			Filename: fields.Filename,
			NumFiles: fields.NumFiles,
		}
	}
	tests := []struct {
		name        string
		fields      fields
		wantErr     bool
		wantRw      *RotateWriter
		startUpFunc func(fields fields) *RotateWriter
		cleanUp     func()
		afterTest   func(t *testing.T)
	}{
		{
			name:    "Wrong filename(e.g. directory)",
			fields:  fields{Filename: "./logs"},
			wantErr: true,
			wantRw:  nil,
			cleanUp: cleanupLogs,
		},
		{
			name:    "Normal case",
			fields:  fields{Filename: "./logs/test.log"},
			wantErr: false,
			wantRw:  &RotateWriter{Filename: "./logs/test.log"},
			cleanUp: cleanupLogs,
			afterTest: func(t *testing.T) {
				_, err := os.Stat("logs/test.log")
				if err != nil {
					t.Errorf("RotateWriter.openWriteFile() error checking file after open: %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rw *RotateWriter
			rw = nil
			if tt.startUpFunc == nil {
				rw = defStartUpfunc(tt.fields)
			} else {
				rw = tt.startUpFunc(tt.fields)
			}
			err := rw.openWriteFile()
			if (err != nil) != tt.wantErr {
				t.Errorf("RotateWriter.openWriteFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err == nil) && !tt.wantErr {
				if !compareRws(rw, tt.wantRw) {
					t.Errorf("RotateWriter.openWriteFile() = %v, want %v", rw, tt.wantRw)
				}
			}
			if tt.afterTest != nil {
				tt.afterTest(t)
			}
			if tt.cleanUp != nil {
				tt.cleanUp()
			}
		})
	}
}

func TestRotateWriter_openWriteFileInt(t *testing.T) {
	type fields struct {
		Filename string
		NumFiles int
		dirpath  string
	}
	defStartUpfunc := func(fields fields) *RotateWriter {
		return &RotateWriter{
			Filename: fields.Filename,
			NumFiles: fields.NumFiles,
		}
	}
	tests := []struct {
		name        string
		fields      fields
		wantFile    bool
		wantErr     bool
		wantRw      *RotateWriter
		startUpFunc func(fields fields) *RotateWriter
		cleanUp     func()
		afterTest   func(t *testing.T)
	}{
		{
			name:     "Wrong filename(e.g. directory)",
			fields:   fields{Filename: "./logs"},
			wantFile: false,
			wantErr:  true,
			wantRw:   nil,
			cleanUp:  cleanupLogs,
		},
		{
			name:     "Normal case",
			fields:   fields{Filename: "./logs/test.log"},
			wantFile: true,
			wantErr:  false,
			wantRw:   &RotateWriter{Filename: "./logs/test.log"},
			cleanUp:  cleanupLogs,
			afterTest: func(t *testing.T) {
				_, err := os.Stat("logs/test.log")
				if err != nil {
					t.Errorf("RotateWriter.openWriteFileInt() error checking file after open: %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rw *RotateWriter
			rw = nil
			if tt.startUpFunc == nil {
				rw = defStartUpfunc(tt.fields)
			} else {
				rw = tt.startUpFunc(tt.fields)
			}
			gotFile, err := rw.openWriteFileInt()
			if (err != nil) != tt.wantErr {
				t.Errorf("RotateWriter.openWriteFileInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && tt.wantFile && (nil == gotFile) {
				t.Errorf("RotateWriter.openWriteFileInt() file was not opened or function does not return them")
			}
			if (err == nil) && !tt.wantErr {
				if !compareRws(tt.wantRw, rw) {
					t.Errorf("RotateWriter.openWriteFileInt() = %v, want %v", rw, tt.wantRw)
				}
			}
			if tt.afterTest != nil {
				tt.afterTest(t)
			}
			if tt.cleanUp != nil {
				tt.cleanUp()
			}
		})
	}
}

func TestRotateWriter_CloseWriteFile(t *testing.T) {
	type fields struct {
		Filename string
		NumFiles int
		dirpath  string
	}
	defStartUpfunc := func(fields fields) *RotateWriter {
		return &RotateWriter{
			Filename: fields.Filename,
			NumFiles: fields.NumFiles,
		}
	}
	tests := []struct {
		name        string
		fields      fields
		wantErr     bool
		startUpFunc func(fields fields) *RotateWriter
		cleanUp     func()
		afterTest   func(t *testing.T)
	}{
		{
			name: "file not opened yet",
			fields: fields{
				Filename: "./logs/test.log",
			},
			wantErr: false,
			cleanUp: cleanupLogs,
		},
		{
			name: "normal writer init(teoretically)",
			fields: fields{
				Filename: "./logs/test.log",
			},
			wantErr: false,
			startUpFunc: func(fields fields) *RotateWriter {
				rw, err := NewRotateWriter(fields.Filename, 0)
				if err != nil {
					t.Errorf("RotateWriter.CloseWriteFile() error during test startup while NewRotateWriter() called: %v", err)
					return nil
				}
				return rw
			},
			cleanUp: cleanupLogs,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rw *RotateWriter
			rw = nil
			if tt.startUpFunc == nil {
				rw = defStartUpfunc(tt.fields)
			} else {
				rw = tt.startUpFunc(tt.fields)
			}
			if rw == nil {
				t.Errorf("RotateWriter.CloseWriteFile() error during test startup occured")
			}
			err := rw.CloseWriteFile()
			if (err != nil) != tt.wantErr {
				t.Errorf("RotateWriter.CloseWriteFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.afterTest != nil {
				tt.afterTest(t)
			}
			if tt.cleanUp != nil {
				tt.cleanUp()
			}
		})
	}
}

func TestRotateWriter_Write(t *testing.T) {
	type fields struct {
		Filename string
		NumFiles int
		dirpath  string
	}
	type args struct {
		p []byte
	}
	defStartUpfunc := func(fields fields) *RotateWriter {
		return &RotateWriter{
			Filename: fields.Filename,
			NumFiles: fields.NumFiles,
		}
	}
	cleanupFunc := func(rw *RotateWriter) {
		rw.CloseWriteFile()
		cleanupLogs()
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantN       int
		wantErr     bool
		startUpFunc func(fields fields) *RotateWriter
		cleanUp     func(rw *RotateWriter)
		afterTest   func(rw *RotateWriter, t *testing.T)
	}{
		{
			name:        "Nil file descriptor when no init or file closed",
			fields:      fields{Filename: "./logs/test.log"},
			args:        args{p: []byte("test string")},
			wantN:       0,
			wantErr:     true,
			startUpFunc: defStartUpfunc,
			cleanUp:     cleanupFunc,
		},
		{
			name:    "Test right file write",
			fields:  fields{Filename: "./logs/test.log"},
			args:    args{p: []byte("test string")},
			wantN:   11,
			wantErr: false,
			startUpFunc: func(fields fields) *RotateWriter {
				cleanupLogs()
				rw, err := NewRotateWriter(fields.Filename, 0)
				if err != nil {
					t.Errorf("RotateWriter.Write() error during test startup while NewRotateWriter() called: %v", err)
					return nil
				}
				return rw
			},
			cleanUp: cleanupFunc,
			afterTest: func(rw *RotateWriter, t *testing.T) {
				if rw == nil {
					return
				}
				rw.CloseWriteFile()
				st, err := os.Stat("./logs/test.log")
				if err != nil {
					t.Errorf("RotateWriter.Write() error while checking what is written to log file %v", err)
					return
				}
				if st.Size() != 11 {
					t.Errorf("RotateWriter.Write() wrong size was written")
					return
				}
				// now read file and compare
				b, err := ioutil.ReadFile("./logs/test.log")
				if err != nil {
					t.Errorf("RotateWriter.Write() error while checking what is written to log file %v", err)
					return
				}
				if !reflect.DeepEqual([]byte("test string"), b) {
					t.Errorf("RotateWriter.Write() error while checking what is written to log file. \nWritten: %v\n, retrieved %v", []byte("test string"), b)
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rw *RotateWriter
			rw = nil
			if tt.startUpFunc == nil {
				rw = defStartUpfunc(tt.fields)
			} else {
				rw = tt.startUpFunc(tt.fields)
			}
			if rw == nil {
				t.Errorf("RotateWriter.Write() error during test startup occured")
			}
			gotN, err := rw.Write(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("RotateWriter.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("RotateWriter.Write() = %v, want %v", gotN, tt.wantN)
			}
			if tt.afterTest != nil {
				tt.afterTest(rw, t)
			}
			if tt.cleanUp != nil {
				tt.cleanUp(rw)
			}
		})
	}
}

func TestRotateWriter_RotationInProgress(t *testing.T) {
	type fields struct {
		Filename string
		NumFiles int
		dirpath  string
	}
	defStartUpfunc := func(fields fields) *RotateWriter {
		return &RotateWriter{
			Filename: fields.Filename,
			NumFiles: fields.NumFiles,
		}
	}
	cleanupFunc := func(rw *RotateWriter) {
		rw.CloseWriteFile()
		cleanupLogs()
	}
	tests := []struct {
		name        string
		fields      fields
		want        bool
		startUpFunc func(fields fields) *RotateWriter
		cleanUp     func(rw *RotateWriter)
		afterTest   func(rw *RotateWriter, t *testing.T)
	}{
		{
			name:   "Just simple case to call that.",
			fields: fields{Filename: "logs/test.log"},
			want:   false,
			startUpFunc: func(fields fields) *RotateWriter {
				cleanupLogs()
				rw, err := NewRotateWriter(fields.Filename, 0)
				if err != nil {
					t.Errorf("RotateWriter.RotationInProgress() error during test startup while NewRotateWriter() called: %v", err)
					return nil
				}
				return rw
			},
			cleanUp: cleanupFunc,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rw *RotateWriter
			rw = nil
			if tt.startUpFunc == nil {
				rw = defStartUpfunc(tt.fields)
			} else {
				rw = tt.startUpFunc(tt.fields)
			}
			if rw == nil {
				t.Errorf("RotateWriter.RotationInProgress() error during test startup occured")
			}
			if got := rw.RotationInProgress(); got != tt.want {
				t.Errorf("RotateWriter.RotationInProgress() = %v, want %v", got, tt.want)
			}
			if tt.afterTest != nil {
				tt.afterTest(rw, t)
			}
			if tt.cleanUp != nil {
				tt.cleanUp(rw)
			}
		})
	}
}

func TestRotateWriter_Rotate(t *testing.T) {
	type fields struct {
		Filename string
		NumFiles int
		dirpath  string
	}
	defStartUpfunc := func(fields fields) *RotateWriter {
		return &RotateWriter{
			Filename: fields.Filename,
			NumFiles: fields.NumFiles,
		}
	}
	cleanupFunc := func(rw *RotateWriter) {
		rw.CloseWriteFile()
		cleanupLogs()
	}
	type args struct {
		ready func()
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		startUpFunc func(fields fields) *RotateWriter
		cleanUp     func(rw *RotateWriter)
		afterTest   func(rw *RotateWriter, t *testing.T)
	}{
		{
			name:    "Write, rotate,write again, check files. here just one test cause function need manual testing in other cases",
			fields:  fields{Filename: "logs/test.log", NumFiles: 2},
			args:    args{ready: nil},
			wantErr: false,
			startUpFunc: func(fields fields) *RotateWriter {
				cleanupLogs()
				rw, err := NewRotateWriter(fields.Filename, fields.NumFiles)
				if err != nil {
					t.Errorf("RotateWriter.Rotate() error during test startup while NewRotateWriter() called: %v", err)
					return nil
				}
				_, err = rw.Write([]byte("test1"))
				if err != nil {
					t.Errorf("RotateWriter.Rotate() error during test startup while RotateWriter.Write() called: %v", err)
					return nil
				}
				return rw
			},
			afterTest: func(rw *RotateWriter, t *testing.T) {
				_, err := rw.Write([]byte("test2"))
				if err != nil {
					t.Errorf("RotateWriter.Rotate() error during test after Rotate run while RotateWriter.Write() called: %v", err)
					return
				}
				rw.CloseWriteFile()
				// stat backed file and test content
				st, err := os.Stat("./logs/test.log.1")
				if err != nil {
					t.Errorf("RotateWriter.Rotate() error while checking what is written to old log file %v", err)
					return
				}
				if st.Size() != 5 {
					t.Errorf("RotateWriter.Rotate() wrong size was written")
					return
				}
				// now read file and compare
				b, err := ioutil.ReadFile("./logs/test.log.1")
				if err != nil {
					t.Errorf("RotateWriter.Rotate() error while checking what is written to old log file %v", err)
					return
				}
				if !reflect.DeepEqual([]byte("test1"), b) {
					t.Errorf("RotateWriter.Rotate() error while checking what is written to old log file. \nWritten: %v\n, retrieved %v", []byte("test1"), b)
					return
				}
				// stat new file and test content
				st, err = os.Stat("./logs/test.log")
				if err != nil {
					t.Errorf("RotateWriter.Rotate() error while checking what is written to new log file %v", err)
					return
				}
				if st.Size() != 5 {
					t.Errorf("RotateWriter.Rotate() wrong size was written")
					return
				}
				// now read file and compare
				b, err = ioutil.ReadFile("./logs/test.log")
				if err != nil {
					t.Errorf("RotateWriter.Rotate() error while checking what is written to new log file %v", err)
					return
				}
				if !reflect.DeepEqual([]byte("test2"), b) {
					t.Errorf("RotateWriter.Rotate() error while checking what is written to new log file. \nWritten: %v\n, retrieved %v", []byte("test2"), b)
					return
				}
			},
			cleanUp: cleanupFunc,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rw *RotateWriter
			rw = nil
			if tt.startUpFunc == nil {
				rw = defStartUpfunc(tt.fields)
			} else {
				rw = tt.startUpFunc(tt.fields)
			}
			if rw == nil {
				t.Errorf("RotateWriter.RotationInProgress() error during test startup occured")
			}
			if err := rw.Rotate(tt.args.ready); (err != nil) != tt.wantErr {
				t.Errorf("RotateWriter.Rotate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.afterTest != nil {
				tt.afterTest(rw, t)
			}
			if tt.cleanUp != nil {
				tt.cleanUp(rw)
			}
		})
	}
}
