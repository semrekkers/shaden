package wav_test

import (
	"fmt"
	"log"
	"os"

	"github.com/go-audio/wav"
)

func ExampleDecoder_Duration() {
	f, err := os.Open("fixtures/kick.wav")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	dur, err := wav.NewDecoder(f).Duration()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s duration: %s\n", f.Name(), dur)
	// Output: fixtures/kick.wav duration: 204.172335ms
}

func ExampleDecoder_IsValidFile() {
	f, err := os.Open("fixtures/kick.wav")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fmt.Printf("is this file valid: %t", wav.NewDecoder(f).IsValidFile())
	// Output: is this file valid: true
}

func ExampleEncoder_Write() {
	f, err := os.Open("fixtures/kick.wav")
	if err != nil {
		panic(fmt.Sprintf("couldn't open audio file - %v", err))
	}

	// Decode the original audio file
	// and collect audio content and information.
	d := wav.NewDecoder(f)
	buf, err := d.FullPCMBuffer()
	if err != nil {
		panic(err)
	}
	f.Close()
	fmt.Println("Old file ->", d)

	// Destination file
	out, err := os.Create("testOutput/kick.wav")
	if err != nil {
		panic(fmt.Sprintf("couldn't create output file - %v", err))
	}

	// setup the encoder and write all the frames
	e := wav.NewEncoder(out,
		buf.Format.SampleRate,
		int(d.BitDepth),
		buf.Format.NumChannels,
		int(d.WavAudioFormat))
	if err := e.Write(buf); err != nil {
		panic(err)
	}
	// close the encoder to make sure the headers are properly
	// set and the data is flushed.
	if err := e.Close(); err != nil {
		panic(err)
	}
	out.Close()

	// reopen to confirm things worked well
	out, err = os.Open("testOutput/kick.wav")
	if err != nil {
		panic(err)
	}
	d2 := wav.NewDecoder(out)
	d2.ReadInfo()
	fmt.Println("New file ->", d2)
	out.Close()
	os.Remove(out.Name())

	// Output:
	// Old file -> Format: WAVE - 1 channels @ 22050 / 16 bits - Duration: 0.204172 seconds
	// New file -> Format: WAVE - 1 channels @ 22050 / 16 bits - Duration: 0.204172 seconds
}

func ExampleDecoder_ReadMetadata() {
	f, err := os.Open("fixtures/listinfo.wav")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	d := wav.NewDecoder(f)
	d.ReadMetadata()
	if d.Err() != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v\n", d.Metadata)
	// Output:
	// &wav.Metadata{SamplerInfo:(*wav.SamplerInfo)(nil), Artist:"artist", Comments:"my comment", Copyright:"", CreationDate:"2017", Engineer:"", Technician:"", Genre:"genre", Keywords:"", Medium:"", Title:"track title", Product:"album title", Subject:"", Software:"", Source:"", Location:"", TrackNbr:"42"}
}
