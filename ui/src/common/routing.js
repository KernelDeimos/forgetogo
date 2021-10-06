import React from "react";

export class RouterState {
    path = ""

    constructor (path) {
        this.path = path;
    }
}

export class Route extends React.Component {
    constructor (props) {
        super(props)
    }
    matches () {
        let routerPath = this.props.rs.path;
        let myPath = this.props.path;
        return routerPath.startsWith(myPath);
    }
    render () {
        return this.matches() ? (
            this.props.children
        ) : [];
    }
}