import React from "react";
import { RouterState } from "./routing";

export function newrs (s) {
    return new RouterState(s);
}

export class BaseComponent extends React.Component {
    render () {
        return []
    }
}