import React from 'react';
import {Chart} from './Chart';
import DataTable from 'react-data-table-component';

class DisplayData extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            error: null,
            isLoaded: false,
            intervalID: null,
            sensors: {
                humidity: [],
                barometer: [],
                gyroscope: [],
                vibration: [],
                light: [],
                gps: []
            }
        };
    }

    componentDidMount() {
        this.intervalID = setInterval(() => this.updateData(), 5000);
    }

    componentWillUnmount() {
        clearInterval(this.intervalID);
    }

    updateData() {
        fetch("/api/channels/common/chaincodes/hlf_iot_cc?fcn=listIotHumidity&args=&peer=hlfiot/peer0")
            .then(res => res.json())
            .then(
                (result) => {
                    this.setState(prevState => ({
                        ...prevState,
                        isLoaded: true,
                        sensors: {
                            ...prevState.sensors,
                            humidity: result.result
                        }
                    }));
                },
                (error) => {
                    this.setState({
                        isLoaded: true,
                        error
                    });
                }
            );

        fetch("/api/channels/common/chaincodes/hlf_iot_cc?fcn=listIotGps&args=&peer=hlfiot/peer0")
            .then(res => res.json())
            .then(
                (result) => {
                    this.setState(prevState => ({
                        ...prevState,
                        isLoaded: true,
                        sensors: {
                            ...prevState.sensors,
                            gps: result.result
                        }
                    }));
                },
                (error) => {
                    this.setState({
                        isLoaded: true,
                        error
                    });
                }
            );

        fetch("/api/channels/common/chaincodes/hlf_iot_cc?fcn=listIotBarometer&args=&peer=hlfiot/peer0")
            .then(res => res.json())
            .then(
                (result) => {
                    this.setState(prevState => ({
                        ...prevState,
                        isLoaded: true,
                        sensors: {
                            ...prevState.sensors,
                            barometer: result.result
                        }
                    }));
                },
                (error) => {
                    this.setState({
                        isLoaded: true,
                        error
                    });
                }
            );

        fetch("/api/channels/common/chaincodes/hlf_iot_cc?fcn=listIotGyroscope&args=&peer=hlfiot/peer0")
            .then(res => res.json())
            .then(
                (result) => {
                    this.setState(prevState => ({
                        ...prevState,
                        isLoaded: true,
                        sensors: {
                            ...prevState.sensors,
                            gyroscope: result.result
                        }
                    }));
                },
                (error) => {
                    this.setState({
                        isLoaded: true,
                        error
                    });
                }
            );

        fetch("/api/channels/common/chaincodes/hlf_iot_cc?fcn=listIotVibration&args=&peer=hlfiot/peer0")
            .then(res => res.json())
            .then(
                (result) => {
                    this.setState(prevState => ({
                        ...prevState,
                        isLoaded: true,
                        sensors: {
                            ...prevState.sensors,
                            vibration: result.result
                        }
                    }));
                },
                (error) => {
                    this.setState({
                        isLoaded: true,
                        error
                    });
                }
            );

        fetch("/api/channels/common/chaincodes/hlf_iot_cc?fcn=listIotLight&args=&peer=hlfiot/peer0")
            .then(res => res.json())
            .then(
                (result) => {
                    this.setState(prevState => ({
                        ...prevState,
                        isLoaded: true,
                        sensors: {
                            ...prevState.sensors,
                            light: result.result
                        }
                    }));
                },
                (error) => {
                    this.setState({
                        isLoaded: true,
                        error
                    });
                }
            );
    }

    render() {
        const lines = {
            humidity: [
                {
                    name: "Humidity",
                    key: "value.humidity",
                    color: "red"
                },
                {
                    name: "Temperature",
                    key: "value.temperature",
                    color: "green"
                }
            ],
            light: [
                {
                    name: "Light",
                    key: "value.light",
                    color: "red"
                }
            ],
            vibration: [
                {
                    name: "Vibration",
                    key: "value.vibration",
                    color: "red"
                }
            ],
            gps: [
                {
                    name: "Longitude",
                    key: "value.longitude",
                    color: "red"
                },
                {
                    name: "Latitude",
                    key: "value.latitude",
                    color: "green"
                },
                {
                    name: "Altitude",
                    key: "value.altitude",
                    color: "blue"
                }
            ],
            barometer: [
                {
                    name: "Pressure",
                    key: "value.pressure",
                    color: "red"
                },
                {
                    name: "Altitude",
                    key: "value.altitude",
                    color: "green"
                },
                {
                    name: "Temperature",
                    key: "value.temperature",
                    color: "blue"
                }
            ],
            gyroscope: [
                {
                    name: "Xout",
                    key: "value.xout",
                    color: "red"
                },
                {
                    name: "Xoutscaled",
                    key: "value.xoutscaled",
                    color: "green"
                },
                {
                    name: "Yout",
                    key: "value.yout",
                    color: "blue"
                },
                {
                    name: "Youtscaled",
                    key: "value.youtscaled",
                    color: "AntiqueWhite"
                },
                {
                    name: "zout",
                    key: "value.zout",
                    color: "Aqua"
                },
                {
                    name: "zoutscaled",
                    key: "value.zoutscaled",
                    color: "black"
                },
                {
                    name: "accelerationxout",
                    key: "value.accelerationxout",
                    color: "coral"
                },
                {
                    name: "accelerationxoutscaled",
                    key: "value.accelerationxoutscaled",
                    color: "DarkGreen"
                },
                {
                    name: "accelerationyout",
                    key: "value.accelerationyout",
                    color: "DarkMagenta"
                },
                {
                    name: "accelerationyoutscaled",
                    key: "value.accelerationyoutscaled",
                    color: "DimGray"
                },
                {
                    name: "accelerationZout",
                    key: "value.accelerationZout",
                    color: "Gold"
                },
                {
                    name: "accelerationZoutscaled",
                    key: "value.accelerationZoutscaled",
                    color: "MediumPurple"
                }
            ],

        };

        const columns = [
            {
                name: 'Timestamp',
                selector: 'timestamp',
                sortable: true,
                cell: row => (new Date(row.value.timestamp * 1000)).toLocaleString()
            },
            {
                name: 'Data',
                selector: 'value',
                cell: row => <div style={{textAlign: 'left'}}>{Object.keys(row.value).map((data, i) => <div
                    key={i}>{data}: {row.value[data]}</div>)}</div>
            },
        ];

        const {error, isLoaded, sensors} = this.state;
        if (error) {
            return <div>Error: {error.message}</div>;
        } else if (!isLoaded) {
            return <div>Loading...</div>;
        } else {
            let sensorsValid = {
                "humidity": [],
                "light": [],
                "vibration": [],
                "gps": [],
                "barometer": [],
                "gyroscope": []
            }, sensorsInvalid = {
                "humidity": [],
                "light": [],
                "vibration": [],
                "gps": [],
                "barometer": [],
                "gyroscope": []
            };

            Object.keys(sensors).forEach((field) => {
                if (sensors[field].length > 0) {
                    sensors[field].forEach((v) => {
                        if (v.value.valid == "1") {
                            sensorsValid[field].push(v)
                        } else {
                            sensorsInvalid[field].push(v)
                        }
                    });
                }
            });

            let sensorsValidData = Object.keys(sensorsValid).map((key, index) => (
                <div>
                    <Chart elements={sensorsValid[key]} lines={lines[key]} keyChart={key} key={key}/>
                    <DataTable
                        title={key + ' data Valid'}
                        columns={columns}
                        data={sensorsValid[key]}
                        pagination={true}
                        defaultSortField={'timestamp'}
                    />
                </div>
            ));

            let sensorsInvalidData = Object.keys(sensorsInvalid).map((key, index) => (
                <div>
                    <DataTable
                        title={key + ' data Invalid'}
                        columns={columns}
                        data={sensorsInvalid[key]}
                        pagination={true}
                        defaultSortField={'timestamp'}
                    />
                </div>
            ));

            return (
                <div>
                    {sensorsValidData}
                    {sensorsInvalidData}
                </div>
            );
        }
    }
}

export {DisplayData as HlfIot};
