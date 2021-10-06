import React from "react";

export const TabContext = React.createContext({
    tabs: [],
    activeTab: null,
});

export class Tab {
    constructor (id, label, elem) {
        this.id = id;
        this.label = label;
        this.elem = elem;
    }
}

export class TabProvider extends React.Component {
    constructor (props) {
        super(props);
        this.state = {
            tabs: props.tabs,
            activeTab: props.activeTab,
            addTab: this.addTab.bind(this),
            switchTab: this.switchTab.bind(this),
        };
    }
    addTab (id, label, elem, noSwitch) {
        for ( let tab of this.state.tabs ) {
            if ( tab.id == id ) {
                this.setState({
                    activeTab: id,
                })
                // TODO: maybe change label also
                return;
            }
        } 
        this.setState({
            tabs: [
                ...this.state.tabs,
                new Tab(id, label, elem)
            ],
            ...(noSwitch ? {} : { activeTab: id }),
        });
    }
    switchTab (id) {
        // TODO: maybe verify that the tab exists
        this.setState({
            activeTab: id,
        });
    }
    render () {
        return (
            <TabContext.Provider value={this.state}>{this.props.children}</TabContext.Provider>
        );
    }
}

export class Tabs extends React.Component {
    render () {
        return (
            <div class="common-tabs">
                <TabContext.Consumer>{tabCtx => {
                    let tabs = tabCtx.tabs.map(tab => {
                        let cls = 'common-tab';
                        let isActive = false;
                        if ( tab.id == tabCtx.activeTab ) {
                            cls += ' common-tab-active';
                            isActive = true;
                        }
                        return (
                            <button
                                className={cls}
                                onClick={isActive
                                    ? () => {} : () => {
                                        tabCtx.switchTab(tab.id);
                                    }
                                }
                            >{tab.label}</button>
                        )
                    });
                    return tabs;
                }}</TabContext.Consumer>
            </div>
        )
    }
}

export class TabContentsView extends React.Component {
    render () {
        return (
            <TabContext.Consumer>{tabCtx => {
                let activeTab = tabCtx.tabs.find(
                    tab => tab.id == tabCtx.activeTab);
                if ( ! activeTab ) return [];
                return activeTab.elem;
            }}</TabContext.Consumer>
        );
    }
}