/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package util

import (
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sort"
	"testing"
)

func BenchmarkParseVersionRule(b *testing.B) {
	f := ParseVersionRule("latest")
	kvs := []*mvccpb.KeyValue{
		{
			Key:   []byte("/service/ver/1.0.300"),
			Value: []byte("1.0.300"),
		},
		{
			Key:   []byte("/service/ver/1.0.303"),
			Value: []byte("1.0.303"),
		},
		{
			Key:   []byte("/service/ver/1.0.304"),
			Value: []byte("1.0.304"),
		},
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f(kvs)
		}
	})
	b.ReportAllocs()
}

var _ = Describe("Version Rule sorter", func() {
	Describe("Sorter", func() {
		Context("Normal", func() {
			It("version asc", func() {
				kvs := []string{"1.0.0", "1.0.1"}
				sort.Sort(&serviceKeySorter{
					sortArr: kvs,
					kvs:     make(map[string]*mvccpb.KeyValue),
					cmp:     Larger,
				})
				Expect(kvs[0]).To(Equal("1.0.1"))
				Expect(kvs[1]).To(Equal("1.0.0"))
			})
			It("version desc", func() {
				kvs := []string{"1.0.1", "1.0.0"}
				sort.Sort(&serviceKeySorter{
					sortArr: kvs,
					kvs:     make(map[string]*mvccpb.KeyValue),
					cmp:     Larger,
				})
				Expect(kvs[0]).To(Equal("1.0.1"))
				Expect(kvs[1]).To(Equal("1.0.0"))
			})
			It("len(v1) != len(v2)", func() {
				kvs := []string{"1.0.0.0", "1.0.1"}
				sort.Sort(&serviceKeySorter{
					sortArr: kvs,
					kvs:     make(map[string]*mvccpb.KeyValue),
					cmp:     Larger,
				})
				Expect(kvs[0]).To(Equal("1.0.1"))
				Expect(kvs[1]).To(Equal("1.0.0.0"))
			})
			It("1.0.9 vs 1.0.10", func() {
				kvs := []string{"1.0.9", "1.0.10"}
				sort.Sort(&serviceKeySorter{
					sortArr: kvs,
					kvs:     make(map[string]*mvccpb.KeyValue),
					cmp:     Larger,
				})
				Expect(kvs[0]).To(Equal("1.0.10"))
				Expect(kvs[1]).To(Equal("1.0.9"))
			})
			It("1.10 vs 4", func() {
				kvs := []string{"1.10", "4"}
				sort.Sort(&serviceKeySorter{
					sortArr: kvs,
					kvs:     make(map[string]*mvccpb.KeyValue),
					cmp:     Larger,
				})
				Expect(kvs[0]).To(Equal("4"))
				Expect(kvs[1]).To(Equal("1.10"))
			})
		})
		Context("Exception", func() {
			It("invalid version1", func() {
				kvs := []string{"1.a", "1.0.1.a", ""}
				sort.Sort(&serviceKeySorter{
					sortArr: kvs,
					kvs:     make(map[string]*mvccpb.KeyValue),
					cmp:     Larger,
				})
				Expect(kvs[0]).To(Equal("1.a"))
				Expect(kvs[1]).To(Equal("1.0.1.a"))
				Expect(kvs[2]).To(Equal(""))
			})
			It("invalid version2 > 32767", func() {
				kvs := []string{"1.0", "1.0.1.32768"}
				sort.Sort(&serviceKeySorter{
					sortArr: kvs,
					kvs:     make(map[string]*mvccpb.KeyValue),
					cmp:     Larger,
				})
				Expect(kvs[0]).To(Equal("1.0"))
				Expect(kvs[1]).To(Equal("1.0.1.32768"))
				kvs = []string{"1.0", "1.0.1.32767"}
				sort.Sort(&serviceKeySorter{
					sortArr: kvs,
					kvs:     make(map[string]*mvccpb.KeyValue),
					cmp:     Larger,
				})
				Expect(kvs[0]).To(Equal("1.0.1.32767"))
				Expect(kvs[1]).To(Equal("1.0"))
			})
		})
	})
	Describe("VersionRule", func() {
		const count = 10
		var kvs = [count]*mvccpb.KeyValue{}
		BeforeEach(func() {
			for i := 1; i <= count; i++ {
				kvs[i-1] = &mvccpb.KeyValue{
					Key:   []byte(fmt.Sprintf("/service/ver/1.%d", i)),
					Value: []byte(fmt.Sprintf("%d", i)),
				}
			}
		})
		Context("Normal", func() {
			It("Latest", func() {
				results := VersionRule(Latest).Match(kvs[:])
				Expect(len(results)).To(Equal(1))
				Expect(results[0]).To(Equal(fmt.Sprintf("%d", count)))
			})
			It("Range1.1 ver in [1.4, 1.8)", func() {
				results := VersionRule(Range).Match(kvs[:], "1.4", "1.8")
				Expect(len(results)).To(Equal(4))
				Expect(results[0]).To(Equal("7"))
				Expect(results[3]).To(Equal("4"))
			})
			It("Range1.2 ver in (1.8, 1.4]", func() {
				results := VersionRule(Range).Match(kvs[:], "1.8", "1.4")
				Expect(len(results)).To(Equal(4))
				Expect(results[0]).To(Equal("7"))
				Expect(results[3]).To(Equal("4"))
			})
			It("Range2 ver in [1, 2]", func() {
				results := VersionRule(Range).Match(kvs[:], "1", "2")
				Expect(len(results)).To(Equal(10))
				Expect(results[0]).To(Equal("10"))
				Expect(results[9]).To(Equal("1"))
			})
			It("Range3 ver in [1.4.1, 1.9.1]", func() {
				results := VersionRule(Range).Match(kvs[:], "1.4.1", "1.9.1")
				Expect(len(results)).To(Equal(5))
				Expect(results[0]).To(Equal("9"))
				Expect(results[4]).To(Equal("5"))
			})
			It("Range4 ver in [2, 4]", func() {
				results := VersionRule(Range).Match(kvs[:], "2", "4")
				Expect(len(results)).To(Equal(0))
			})
			It("AtLess1 ver >= 1.6", func() {
				results := VersionRule(AtLess).Match(kvs[:], "1.6")
				Expect(len(results)).To(Equal(5))
				Expect(results[0]).To(Equal("10"))
				Expect(results[4]).To(Equal("6"))
			})
			It("AtLess2 ver >= 1", func() {
				results := VersionRule(AtLess).Match(kvs[:], "1")
				Expect(len(results)).To(Equal(10))
				Expect(results[0]).To(Equal("10"))
				Expect(results[9]).To(Equal("1"))
			})
			It("AtLess3 ver >= 1.5.1", func() {
				results := VersionRule(AtLess).Match(kvs[:], "1.5.1")
				Expect(len(results)).To(Equal(5))
				Expect(results[0]).To(Equal("10"))
				Expect(results[4]).To(Equal("6"))
			})
			It("AtLess4 ver >= 2", func() {
				results := VersionRule(AtLess).Match(kvs[:], "2")
				Expect(len(results)).To(Equal(0))
			})
		})
		Context("Exception", func() {
			It("nil", func() {
				results := VersionRule(Latest).Match(nil)
				Expect(len(results)).To(Equal(0))
				results = VersionRule(AtLess).Match(nil)
				Expect(len(results)).To(Equal(0))
				results = VersionRule(Range).Match(nil)
				Expect(len(results)).To(Equal(0))
				Expect(ParseVersionRule("")).To(BeNil())
				Expect(ParseVersionRule("abc")).To(BeNil())
				Expect(VersionMatchRule("1.0", "1.0")).To(BeTrue())
				Expect(VersionMatchRule("1.0", "1.2")).To(BeFalse())
			})
		})
		Context("Parse", func() {
			It("Latest", func() {
				match := ParseVersionRule("latest")
				results := match(kvs[:])
				Expect(len(results)).To(Equal(1))
				Expect(results[0]).To(Equal(fmt.Sprintf("%d", count)))
			})
			It("Range ver in [1.4, 1.8)", func() {
				match := ParseVersionRule("1.4-1.8")
				results := match(kvs[:])
				Expect(len(results)).To(Equal(4))
				Expect(results[0]).To(Equal("7"))
				Expect(results[3]).To(Equal("4"))
			})
			It("AtLess ver >= 1.6", func() {
				match := ParseVersionRule("1.6+")
				results := match(kvs[:])
				Expect(len(results)).To(Equal(5))
				Expect(results[0]).To(Equal("10"))
				Expect(results[4]).To(Equal("6"))
			})
		})
		Context("VersionMatchRule", func() {
			It("Latest", func() {
				Expect(VersionMatchRule("1.0", "latest")).To(BeTrue())
			})
			It("Range ver in [1.4, 1.8]", func() {
				Expect(VersionMatchRule("1.4", "1.4-1.8")).To(BeTrue())
				Expect(VersionMatchRule("1.6", "1.4-1.8")).To(BeTrue())
				Expect(VersionMatchRule("1.8", "1.4-1.8")).To(BeFalse())
				Expect(VersionMatchRule("1.0", "1.4-1.8")).To(BeFalse())
				Expect(VersionMatchRule("1.9", "1.4-1.8")).To(BeFalse())
			})
			It("AtLess ver >= 1.6", func() {
				Expect(VersionMatchRule("1.6", "1.6+")).To(BeTrue())
				Expect(VersionMatchRule("1.9", "1.6+")).To(BeTrue())
				Expect(VersionMatchRule("1.0", "1.6+")).To(BeFalse())
			})
		})
	})
	Describe("NewVersionRegexp", func() {
		Context("Normal", func() {
			It("Latest", func() {
				vr := NewVersionRegexp(false)
				Expect(vr.MatchString("latest")).To(BeFalse())
				vr = NewVersionRegexp(true)
				Expect(vr.MatchString("latest")).To(BeTrue())
			})
			It("Range", func() {
				vr := NewVersionRegexp(false)
				Expect(vr.MatchString("1.1-2.2")).To(BeFalse())
				vr = NewVersionRegexp(true)
				Expect(vr.MatchString("-")).To(BeFalse())
				Expect(vr.MatchString("1.1-")).To(BeFalse())
				Expect(vr.MatchString("-1.1")).To(BeFalse())
				Expect(vr.MatchString("1.a-2.b")).To(BeFalse())
				Expect(vr.MatchString("1.-.2")).To(BeFalse())
				Expect(vr.MatchString("60000-1")).To(BeFalse())
				Expect(vr.MatchString("1.1-2.2")).To(BeTrue())
			})
			It("AtLess", func() {
				vr := NewVersionRegexp(false)
				Expect(vr.MatchString("1.0+")).To(BeFalse())
				vr = NewVersionRegexp(true)
				Expect(vr.MatchString("+")).To(BeFalse())
				Expect(vr.MatchString("+1.0")).To(BeFalse())
				Expect(vr.MatchString("1.a+")).To(BeFalse())
				Expect(vr.MatchString(".1+")).To(BeFalse())
				Expect(vr.MatchString("1.+")).To(BeFalse())
				Expect(vr.MatchString(".+")).To(BeFalse())
				Expect(vr.MatchString("60000+")).To(BeFalse())
				Expect(vr.MatchString("1.0+")).To(BeTrue())
			})
			It("Explicit", func() {
				vr := NewVersionRegexp(false)
				Expect(vr.MatchString("")).To(BeFalse())
				Expect(vr.MatchString("a")).To(BeFalse())
				Expect(vr.MatchString("60000")).To(BeFalse())
				Expect(vr.MatchString(".")).To(BeFalse())
				Expect(vr.MatchString("1.")).To(BeFalse())
				Expect(vr.MatchString(".1")).To(BeFalse())
				Expect(vr.MatchString("1.4")).To(BeTrue())
				vr = NewVersionRegexp(true)
				Expect(vr.MatchString("")).To(BeFalse())
				Expect(vr.MatchString("a")).To(BeFalse())
				Expect(vr.MatchString("60000")).To(BeFalse())
				Expect(vr.MatchString(".")).To(BeFalse())
				Expect(vr.MatchString("1.")).To(BeFalse())
				Expect(vr.MatchString(".1")).To(BeFalse())
				Expect(vr.MatchString("1.4")).To(BeTrue())
			})
		})
	})
})
