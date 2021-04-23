//  Copyright 2018 Google Inc. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package daisy

import (
	"context"
	"fmt"
	"testing"
)

func TestSuspendInstancesPopulate(t *testing.T) {
	w := testWorkflow()
	s, _ := w.NewStep("s")
	s.SuspendInstances = &SuspendInstances{
		Instances: []string{"i", "zones/z/instances/i"},
	}

	if err := (s.SuspendInstances).populate(context.Background(), s); err != nil {
		t.Error("err should be nil")
	}

	want := &SuspendInstances{
		Instances: []string{"i", fmt.Sprintf("projects/%s/zones/z/instances/i", w.Project)},
	}
	if diffRes := diff(s.SuspendInstances, want, 0); diffRes != "" {
		t.Errorf("SuspendInstances not populated as expected: (-got,+want)\n%s", diffRes)
	}
}

func TestSuspendInstancesValidate(t *testing.T) {
	ctx := context.Background()
	// Set up.
	w := testWorkflow()
	s, _ := w.NewStep("s")
	iCreator, _ := w.NewStep("iCreator")
	iCreator.CreateInstances = &CreateInstances{Instances: []*Instance{&Instance{}}}
	w.AddDependency(s, iCreator)
	if err := w.instances.regCreate("instance1", &Resource{link: fmt.Sprintf("projects/%s/zones/%s/disks/d", testProject, testZone)}, false, iCreator); err != nil {
		t.Fatal(err)
	}

	if err := (&SuspendInstances{Instances: []string{"instance1"}}).validate(ctx, s); err != nil {
		t.Errorf("validation should not have failed: %v", err)
	}

	if err := (&SuspendInstances{Instances: []string{"dne"}}).validate(ctx, s); err == nil {
		t.Error("SuspendInstances should have returned an error when suspending an instance that DNE")
	}
}

func TestSuspendInstancesRun(t *testing.T) {
	ctx := context.Background()
	w := testWorkflow()

	s, _ := w.NewStep("s")
	ins := []*Resource{{RealName: "in0", link: "link"}, {RealName: "in1", link: "link"}}
	w.instances.m = map[string]*Resource{"in0": ins[0], "in1": ins[1]}

	si := &SuspendInstances{
		Instances: []string{"in0"},
	}
	if err := si.run(ctx, s); err != nil {
		t.Fatalf("error running SuspendInstances.run(): %v", err)
	}

	suspendedChecks := []struct {
		r               *Resource
		shouldBeSuspended bool
	}{
		{ins[0], true},
		{ins[1], false},
	}
	for _, c := range suspendedChecks {
		if c.shouldBeSuspended {
			if !c.r.suspendedByWf {
				t.Errorf("resource %q should have been suspended", c.r.RealName)
			}
		} else if c.r.suspendedByWf {
			t.Errorf("resource %q should not have been suspended", c.r.RealName)
		}
	}
}
