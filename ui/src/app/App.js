import React from 'react';
import { BaseComponent, newrs } from '../common/base';
import { Route, RouterState } from '../common/routing';
import { NoConfig, ErrorConfig, CreateConfig, ListConfigs } from '../screens/config';

import '../app.sass'
import '../theme.sass'
import '../tabs.sass'
import { easyFetch, easyPost } from '../common/comm';
import { Tab, TabContentsView, TabProvider, Tabs } from '../common/tabs';
import Konami from 'react-konami-code';

export class App extends BaseComponent {
    constructor (props) {
        super(props);
        this.state = {
            rs: newrs('init.auth')
        };
    } 

    async checkConfig () {
        this.setState({
            rs: newrs('init.loading'),
        });
        let config = await this.doFetch(
            window.location.origin + "/config/launchers");
        if ( ! config ) return;

        if ( config.length < 1 ) {
            this.setState({
                rs: newrs('init.no-config'),
            });
            return;
        }

        let instances = await this.doFetch(
            window.location.origin + "/instance/instances");
        
        this.setState({
            launcherConfigs: config,
            instances: instances,
            rs: newrs('screen.launchers'),
        })
    }

    async openConfigCreate () {
        this.setState({
            rs: newrs('init.config.create'),
        });
    }

    async onCreateConfig (values) {
        const [error, result] = await easyPost(
            window.location.origin + "/config/launchers",
            values,
        );
        if ( error ) {
            // TODO: better to report back to the form
            //       then to set a top-level error state
            this.setState({
                rs: newrs('init.critical-error'),
                errorInfo: result,
            });
            return;
        }

        // Unblocking async is intentional here
        this.checkConfig();
    }

    async doFetch(url) {
        const [error, result] = await easyFetch(url);
        if ( error ) {
            this.setState({
                rs: newrs('init.critical-error'),
                errorInfo: result
            });
            return null;
        }
        return result;
    }

    render () {
        console.log('rendering app')
        return (
            <div class="forgetg-root">
                <Konami action={this.checkConfig.bind(this)}></Konami>
                <Route rs={this.state.rs} path="init.auth">
                    please authenticate
                </Route>
                <Route rs={this.state.rs} path="init.loading">
                    loading
                </Route>
                <Route rs={this.state.rs} path="init.no-config">
                    <NoConfig onCreateConfig={this.openConfigCreate.bind(this)}></NoConfig>
                </Route>
                <Route rs={this.state.rs} path="init.config.create">
                    <CreateConfig onSubmit={this.onCreateConfig.bind(this)}></CreateConfig>
                </Route>
                <Route rs={this.state.rs} path="init.critical-error">
                    <ErrorConfig info={this.state.errorInfo}></ErrorConfig>
                </Route>
                <Route rs={this.state.rs} path="screen">
                    <TabProvider
                        forceUpdateA={this.state.launcherConfigs}
                        forceUpdateB={this.state.instances}
                        tabs={[
                            new Tab('screen.launchers', 'Launch Configs', (
                                <ListConfigs
                                    configs={this.state.launcherConfigs}
                                    instances={this.state.instances}
                                ></ListConfigs>
                            )),
                        ]}
                        activeTab='screen.launchers'
                    >
                        <Tabs></Tabs>
                        <TabContentsView></TabContentsView>
                    </TabProvider>
                </Route>
            </div>
        )
    }
}
