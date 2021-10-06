import Ansi from "ansi-to-react";
import React, { useRef } from "react";
import { easyFetch, easyPost } from "../common/comm";

export class ServerConsole extends React.Component {
    constructor (props) {
        super(props)
        this.state = {
            lines: [],
            scrolling: false,
            commandLine: '',
        };
        this.consoleEndRef = React.createRef();
    }

    componentDidMount() {
        this.init();
        this.updateScrollPosition();
    }

    componentWillUnmount() {
        this.consoleSocket.close();
    }

    componentDidUpdate() {
        this.updateScrollPosition();
    }

    updateScrollPosition () {
        if ( this.state.scrolling ) return;
        let curr = this.consoleEndRef.current;
        if ( ! curr ) {
            console.warn('console end not found');
            return
        }
        curr.scrollIntoView({ behaviour: 'smooth' });
    }
    onConsoleScroll (evt) {
        let t = evt.target;
        let up = t.scrollHeight - t.clientHeight > t.scrollTop;
        if ( ! up ) return;
        this.setState({
            scrolling: true,
        });
    }

    async init () {
        await this.populateBuffer();
        this.initConsoleSocket();
    }
    async populateBuffer() {
        const [error, result] = await easyFetch(
            window.location.origin +
            "/instance/buffer/" +
            this.props.uuid
        )
        if ( error ) {
            // TODO: error handle
            alert('error');
            console.error(result);
            return;
        }

        this.setState({
            lines: result,
        })
    }
    initConsoleSocket () {
        this.consoleSocket = new WebSocket(
            `ws://${window.location.host}/instance/stream/`
            + this.props.uuid,
        );
        this.consoleSocket.onmessage = evt => {
            this.setState({
                // Sure; react is fast, but is this okay?
                lines: [
                    ...this.state.lines,
                    ...('' + evt.data).split('\n')
                ],
            });
        }
    }
    async onCommandEntered (evt) {
        evt.preventDefault();
        let cmd = this.state.commandLine;
        const [error, result] = await easyPost(
            window.location.origin + '/instance/command/' +
            this.props.uuid,
            cmd
        )
        if ( error ) {
            // TODO: error handle
            alert('error running command')
            console.error(result);
            return;
        }
        let cmdLine = `<you entered> ${cmd}`;
        this.setState({
            lines: [...this.state.lines, cmdLine],
        });
    }
    handleCommandLine (evt) {
        this.setState({
            commandLine: evt.target.value,
        });
    }
    render () {
        return (
            <div className="app-serverconsole">
                <div
                    className="app-serverconsole-console"
                    onScroll={this.onConsoleScroll.bind(this)}
                >
                    { this.state.scrolling ? (
                        <button
                            className="app-serverconsole-scrollbtn"
                            onClick={() => {
                                this.setState({
                                    scrolling: false,
                                });
                            }}
                        >
                            Follow Output
                        </button>
                    ) : [] }
                    {this.state.lines.map(line => (
                        <div className="app-serverconsole-line">
                            <Ansi>{line}</Ansi>
                        </div>
                    ))}
                    <div ref={this.consoleEndRef}></div>
                </div>
                <form
                    className="app-serverconsole-entry"
                    onSubmit={this.onCommandEntered.bind(this)}
                >
                    <input
                        name="cmd"
                        value={this.state.commandLine}
                        onChange={this.handleCommandLine.bind(this)}
                    />
                    <input type="submit" value=">" />
                </form>
            </div>
        )
    }
}
