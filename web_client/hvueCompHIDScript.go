// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mame82/hvue"
	"strconv"
)

type CompHIDScriptCodeEditorData struct {
	*js.Object
	CodeMirrorOptions *CodeMirrorOptionsType `js:"codemirrorOptions"`
}

func SendAndRun(vm *hvue.VM) {
	sourceCode := vm.Get("$store").Get("state").Get("currentHIDScriptSource").String()
	md5 := StringToMD5(sourceCode) //Calculate MD5 hexstring of current script content

	go func() {
		timeout := uint32(0)
		err := UploadHIDScript(md5, sourceCode)
		if err != nil {
			QuasarNotifyError("Error uploading script", err.Error(), QUASAR_NOTIFICATION_POSITION_TOP)
			return
		}
		job,err := RunHIDScript(md5, timeout)
		if err != nil {
			QuasarNotifyError("Error starting script as background job", err.Error(), QUASAR_NOTIFICATION_POSITION_TOP)
			return
		}

		QuasarNotifySuccess("Script started successfully", "Job ID " + strconv.Itoa(int(job.Id)), QUASAR_NOTIFICATION_POSITION_TOP)
	}()
}

type CodeMirrorMode struct {
	*js.Object
	Name string `js:"name"`
	GlobalVars bool `js:"globalVars"`
}

type CodeMirrorExtraKeys struct {
	*js.Object
	CtrlSpace string `js:"Ctrl-Space"`
}

type CodeMirrorOptionsType struct {
	*js.Object
	Mode *CodeMirrorMode `js:"mode"`
	LineNumbers bool `js:"lineNumbers"`
	LineWrapping bool `js:"lineWrapping"`
	AutoCloseBrackets bool `js:"autoCloseBrackets"`
	ExtraKeys *CodeMirrorExtraKeys `js:"extraKeys"`
}

func newCompHIDScriptCodeEditorData(vm *hvue.VM) interface{} {
	data := &CompHIDScriptCodeEditorData{ Object: O() }

	data.CodeMirrorOptions = &CodeMirrorOptionsType{Object: O()}

	data.CodeMirrorOptions.Mode = &CodeMirrorMode{ Object: O() }
	data.CodeMirrorOptions.Mode.Name = "text/javascript"
	data.CodeMirrorOptions.Mode.GlobalVars = true //expose globalVars of mode for auto-complete with addon/hint/show-hint.js, addon/hint/javascript-hint.js"

	data.CodeMirrorOptions.ExtraKeys = &CodeMirrorExtraKeys{ Object: O() }
	data.CodeMirrorOptions.ExtraKeys.CtrlSpace = "autocomplete"

	data.CodeMirrorOptions.LineNumbers = true
	//data.CodeMirrorOptions.LineWrapping = true
	data.CodeMirrorOptions.AutoCloseBrackets = true

	return data
}


func InitComponentsHIDScript() {
	hvue.NewComponent(
		"hid-script-code-editor",
		hvue.Template(compHIDScriptCodeEditorTemplate),
		hvue.DataFunc(newCompHIDScriptCodeEditorData),
		hvue.Method("SendAndRun",	SendAndRun),
		hvue.ComputedWithGetSet(
			"scriptContent",
			func(vm *hvue.VM) interface{} {
				return vm.Get("$store").Get("state").Get("currentHIDScriptSource")
			},
			func(vm *hvue.VM, newScriptContent *js.Object) {
				vm.Get("$store").Call("commit", VUEX_MUTATION_SET_CURRENT_HID_SCRIPT_SOURCE_TO, newScriptContent)
			}),
		)

	hvue.NewComponent(
		"hid-script",
		hvue.Template(compHIDScriptTemplate),
		hvue.DataFunc(func(vm *hvue.VM) interface{} {
			data := struct {
				*js.Object
				ShowLoadHIDScriptModal bool   `js:"ShowLoadHIDScriptModal"`
				ShowStoreHIDScriptModal bool   `js:"ShowStoreHIDScriptModal"`
			}{Object: O()}
			data.ShowLoadHIDScriptModal = false
			data.ShowStoreHIDScriptModal = false
			return &data
		}),
		hvue.Method("updateStoredHIDScriptsList",
			func(vm *hvue.VM) {
				vm.Store.Call("dispatch", VUEX_ACTION_UPDATE_STORED_HID_SCRIPTS_LIST)
			}),
		hvue.Method("SendAndRun",	SendAndRun),
	)
}

const (

	compHIDScriptTemplate = `
<q-page padding>
	<modal-string-input v-model="ShowStoreHIDScriptModal" title="Store HIDScript" @save=""></modal-string-input>
	<select-string-from-array :values="this.$store.state.StoredHIDScriptsList" v-model="ShowLoadHIDScriptModal" title="Load HIDScript to editor" @load=""></select-string-from-array>


	<div class="row gutter-sm">

		<div class="col-12">
			<q-card>
  				<q-card-title>
    				HIDScript editor
  				</q-card-title>

				<q-card-main>
					<div class="row gutter-sm">
	    				<div class="col-12 col-sm"><q-btn class="fit" color="primary" label="run" @click="SendAndRun()" icon="play_circle_filled" /></div>
    					<div class="col-12 col-sm"><q-btn class="fit" color="secondary" label="store" icon="save" @click="ShowStoreHIDScriptModal=true" /></div>
    					<div class="col-12 col-sm"><q-btn class="fit" color="warning" label="load & replace" icon="settings_backup_restore" @click="updateStoredHIDScriptsList(); ShowLoadHIDScriptModal=true"/></div>
    					<div class="col-12 col-sm"><q-btn class="fit" color="warning" label="load & prepend" icon="add_to_photos" @click="updateStoredHIDScriptsList(); ShowLoadHIDScriptModal=true"/></div>
    					<!-- <div class="col-12 col-sm"><q-btn class="fit" color="warning" label="load & add" @click="updateStoredTriggerActionSetsList(); showAddTASModal=true" icon="add_to_photos" /></div> -->
					</div>
  				</q-card-main>

			</q-card>
		</div>


		<div class="col-12 col-md-7 col-lg-8 col-xl-9">
			<hid-script-code-editor></hid-script-code-editor>
		</div>
		<div class="col-12 col-md-5 col-lg-4 col-xl-3">
			<hid-job-overview></hid-job-overview>
		</div>
		<div class="col-12" style="overflow: auto; max-height: 40vh;">
			<hid-job-event-overview></hid-job-event-overview>
		</div>
	</div>
</q-page>
`
	compHIDScriptCodeEditorTemplate = `
	<q-card class="full-height">
<!--
  		<q-card-title>
    		HIDScript editor
  		</q-card-title>
-->
		<q-card-main>
			<codemirror v-model="scriptContent" :options="codemirrorOptions"></codemirror>
	  	</q-card-main>
	
	</q-card>
`
)

