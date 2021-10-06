import { Field, Form, Formik } from "formik";
import React from "react";
import { easyFetch, easyPost } from "../common/comm";
import { DyformikForm } from "../common/dyformik";
import { TabContext } from "../common/tabs";
import { ServerConsole } from "./server";

export class NoConfig extends React.Component {
    render () {
        return (
            <div class="app-noconfig">
                <h1>No configuration found!</h1>
                <p>
                    Click the button below to
                    proceed in the working directory.
                </p>
                <input
                    type="button"
                    value="Create Config"
                    onClick={this.props.onCreateConfig}
                ></input>
            </div>
        );
    }
}

export class ErrorConfig extends React.Component {
    render () {
        return (
            <div>
                <h1>Error reading configuration</h1>
                <pre>{JSON.stringify(this.props.info, null, '  ')}</pre>
            </div>
        );
    }
}

// TODO: this could come from the server
const fields = [
    {
        name: 'name',
        // TODO: default label
        label: 'Server Name'
    },
    {
        name: 'jarFile',
        label: 'Jar File',
    },
    {
        name: 'heapMin',
        label: 'Heap (min)',
        type: 'integer',
        initial: 5,
    },
    {
        name: 'heapMax',
        label: 'Heap (max)',
        type: 'integer',
        initial: 5,
    },
    {
        name: 'enableG1GC',
        label: "Enable G1GC",
        type: 'checkbox',
        initial: true
    }
];

export class CreateConfig extends React.Component {
    render () {
        let dyf = new DyformikForm(fields);
        dyf.onSubmit = async values => {
            await this.props.onSubmit(values);
        };
        return (
            <div className="app-createconfig">
                <h1>Create config</h1>
                {dyf.getFormik()}
                {/* <Formik
                    initialValues={fieldToInitial(fields)}
                    onSubmit={this.props.onSubmit}
                >
                    {({ isSubmitting }) =>
                        this.createForm(isSubmitting)}
                </Formik> */}
            </div>
        );
    }
}

export class LaunchConfig extends React.Component {
    constructor (props) {
        super(props);
        this.state = {
            launchDisabled: false,
            stopDisabled: false,
        };
    }
    async launch (addTab) {
        this.setState({
            launchDisabled: true,
        });
        let config = this.props.config
        console.log('config?', config)
        const [error, result] = await easyPost(
            window.location.origin + "/instance/instance/" +
                config.uuid,
            {},
        )
        if ( error ) {
            console.error(result);
            alert('operation failed');
            return;
        }
        this.open(addTab);
        console.log('post instance', result);
    }
    async open (addTab) {
        let config = this.props.config
        addTab(
            'server.' + config.uuid,
            'Server: ' + config.values.name,
            <ServerConsole
                uuid={config.uuid}
            ></ServerConsole>
        );
    }
    async terminate (addTab) {
        this.setState({
            stopDisabled: true,
        })
        let config = this.props.config;
        const [error, result] = await easyPost(
            window.location.origin + '/instance/command/' +
            config.uuid,
            'stop'
        );
        if ( error ) {
            alert('error stopping server');
            console.error(result)
        }
    }
    getControlsForInstance () {
        if ( ! this.props.instance )
            throw new Error('this should only be called if an instance exists');
        let status = this.props.instance.status;

        if ( status == 'running' ) return [
            this.getOpenButton(),
            this.getStopButton('Stop')
        ];
        if ( status == 'crashed' ) return [
            this.getOpenButton(),
            this.getStartButton('Restart')
        ];
        if ( status == 'stopped' ) return [
            this.getOpenButton(),
            this.getStartButton('Restart')
        ];
        return []
    }
    getStartButton (label) {
        return (
            <TabContext.Consumer>{tabCtx => (
                <button
                    className="btn-primary"
                    onClick={this.launch.bind(this, tabCtx.addTab)}
                    disabled={this.state.launchDisabled}
                >{label}</button>
            )}</TabContext.Consumer>
        );
    }
    getStopButton (label) {
        return (
            <TabContext.Consumer>{tabCtx => (
                <button
                    className="btn-primary"
                    onClick={this.terminate.bind(this, tabCtx.addTab)}
                    disabled={this.state.stopDisabled}
                >{label}</button>
            )}</TabContext.Consumer>
        );
    }
    getOpenButton () {
        return (
            <TabContext.Consumer>{tabCtx => (
                <button
                    className="btn-primary"
                    onClick={this.open.bind(this, tabCtx.addTab)}
                >Console</button>
            )}</TabContext.Consumer>
        );
    }
    render () {
        let config = this.props.config;
        return (
            <div className="app-listconfigs-config">
                <div class="app-listconfigs-config-title">
                    {config.values.name}
                </div>
                <div class="app-listconfigs-config-status">{
                    this.props.instance ? this.props.instance.status : 'not running'
                }</div>
                <div class="app-listconfigs-config-controls">
                {
                    this.props.instance
                        ? this.getControlsForInstance()
                        : this.getStartButton('Launch')
                }
                </div>
            </div>
        );
    }
}

export class ListConfigs extends React.Component {
    constructor (props) {
        super(props);
        this.configPoller = null;
        this.state = {
            configs: props.configs,
            instances: props.instances,
            loading: true,
        };
    }
    componentDidMount () {
        this.configPoller = this.startConfigPoller();
    }
    componentWillUnmount () {
        if ( this.configPoller ) clearTimeout(this.configPoller)
    }
    startConfigPoller () {
        this.configPoller = setTimeout(async () => {
            console.log('poll start');
            let [error1, config] = await easyFetch(
                window.location.origin + "/config/launchers");
            if ( error1 ) {
                // todo: error handle
                console.error(config)
                return;
            }
            let [error2, instances] = await easyFetch(
                window.location.origin + "/instance/instances");
            if ( error2 ) {
                // todo: error handle
                console.error(instances)
                return;
            }
            this.setState({
                configs: config,
                instances: instances,
                loading: false,
            })
            console.log('poll end');
            this.startConfigPoller();
        }, 400);
    }
    render () {
        console.log('list configs render');
        if ( this.state.loading ) return (
            <p>Loading...</p>
        );
        return (
            <div className="app-listconfigs">
                <h1>configs</h1>
                { this.state.configs.map(config => {
                    let instance = this.state.instances.find(
                        v => v.uuid == config.uuid)
                    return (
                        <LaunchConfig
                            config={config}
                            instance={instance}
                        ></LaunchConfig>
                    );
                }) }
            </div>
        );
    }
}
