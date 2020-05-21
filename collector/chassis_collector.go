package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	"sync"
	"os"

)

// A ChassisCollector implements the prometheus.Collector.

var (
	ChassisSubsystem = "chassis"
	ChassisLabelNames = []string{"resource", "chassis_id"}
	ChassisTemperatureLabelNames    = []string{"resource", "chassis_id", "sensor", "sensor_id"}
	ChassisFanLabelNames            = []string{"resource", "chassis_id", "fan", "fan_id"}
	ChassisPowerVotageLabelNames    = []string{"resource", "chassis_id", "power_votage", "power_votage_id"}
	ChassisPowerSupplyLabelNames    = []string{"resource", "chassis_id", "power_supply", "power_supply_id"}
	ChassisNetworkAdapterLabelNames = []string{"resource", "chassis_id", "network_adapter", "network_adapter_id"}
	ChassisNetworkPortLabelNames    = []string{"resource", "chassis_id", "network_adapter", "network_adapter_id", "network_port", "network_port_id", "network_port_type", "network_port_speed"}

	chassisMetrics = map[string]chassisMetric{
		"chassis_health": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "health"),
				"health of chassis, 1(OK),2(Warning),3(Critical)",
				ChassisLabelNames,
				nil,
			),
		},
		"chassis_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "state"),
				"state of chassis,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisLabelNames,
				nil,
			),
		},
		"chassis_temperature_sensor_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "temperature_sensor_state"),
				"status state of temperature on this chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisTemperatureLabelNames,
				nil,
			),
		},
		"chassis_temperature_celsius": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "temperature_celsius"),
				"celsius of temperature on this chassis component",
				ChassisTemperatureLabelNames,
				nil,
			),
		},
		"chassis_fan_health": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "fan_health"),
				"fan health on this chassis component,1(OK),2(Warning),3(Critical)",
				ChassisFanLabelNames,
				nil,
			),
		},
		"chassis_fan_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "fan_state"),
				"fan state on this chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisFanLabelNames,
				nil,
			),
		},
		"chassis_fan_rpm": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "fan_rpm_percentage"),
				"fan rpm percentage on this chassis component",
				ChassisFanLabelNames,
				nil,
			),
		},
		"chassis_power_voltage_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_voltage_state"),
				"power voltage state of chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisPowerVotageLabelNames,
				nil,
			),
		},
		"chassis_power_voltage_volts": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_voltage_volts"),
				"power voltage volts number of chassis component",
				ChassisPowerVotageLabelNames,
				nil,
			),
		},
		"chassis_power_powersupply_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_powersupply_state"),
				"powersupply state of chassis component,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisPowerSupplyLabelNames,
				nil,
			),
		},
		"chassis_power_powersupply_health": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_powersupply_health"),
				"powersupply health of chassis component,1(OK),2(Warning),3(Critical)",
				ChassisPowerSupplyLabelNames,
				nil,
			),
		},
		"chassis_power_powersupply_last_power_output_watts": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_powersupply_last_power_output_watts"),
				"last_power_output_watts of powersupply on this chassis",
				ChassisPowerSupplyLabelNames,
				nil,
			),
		},
		"chassis_power_powersupply_power_capacity_watts": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_powersupply_power_capacity_watts"),
				"power_capacity_watts of powersupply on this chassis",
				ChassisPowerSupplyLabelNames,
				nil,
			),
		},
		"chassis_network_adapter_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "network_adapter_state"),
				"chassis network adapter state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisNetworkAdapterLabelNames,
				nil,
			),
		},
		"chassis_network_adapter_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "network_adapter_health_state"),
				"chassis network adapter health state,1(OK),2(Warning),3(Critical)",
				ChassisNetworkAdapterLabelNames,
				nil,
			),
		},
		"chassis_network_port_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "network_port_state"),
				"chassis network port state state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				ChassisNetworkPortLabelNames,
				nil,
			),
		},
		"chassis_network_port_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "network_port_health_state"),
				"chassis network port state state,1(OK),2(Warning),3(Critical)",
				ChassisNetworkPortLabelNames,
				nil,
			),
		},
		"chassis_power_powercontrol_powerallocatedwatts": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, ChassisSubsystem, "power_allocated_watts"),
				"heres testing some stuff lol",
				ChassisPowerSupplyLabelNames,
				nil,
			),
		},
	}
)

type ChassisCollector struct {
	redfishClient           *gofish.APIClient
	metrics                 map[string]chassisMetric
	collectorScrapeStatus   *prometheus.GaugeVec
}

type chassisMetric struct {
	desc *prometheus.Desc
}

// NewChassisCollector returns a collector that collecting chassis statistics
func NewChassisCollector(namespace string, redfishClient *gofish.APIClient) *ChassisCollector {

	// get service from redfish client

	return &ChassisCollector{
		redfishClient: redfishClient,
		metrics:       chassisMetrics,
		collectorScrapeStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "collector_scrape_status",
				Help:      "collector_scrape_status",
			},
			[]string{"collector"},
		),
	}
}

func (c *ChassisCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.desc
	}
	c.collectorScrapeStatus.Describe(ch)

}

func (c *ChassisCollector) Collect(ch chan<- prometheus.Metric) {

	service := c.redfishClient.Service

	// get a list of chassis from service
	if chassises, err := service.Chassis(); err != nil {
		log.Infof("Errors Getting chassis from service : %s", err)
	} else {
		// process the chassises
		for _, chassis := range chassises {
			chassisID := chassis.ID
			//os.Stderr.WriteString(chassis.ID)
			chassisStatus := chassis.Status
			chassisStatusState := chassisStatus.State
			chassisStatusHealth := chassisStatus.Health
			ChassisLabelValues := []string{"chassis", chassisID}
			if chassisStatusHealthValue, ok := parseCommonStatusHealth(chassisStatusHealth); ok {
				ch <- prometheus.MustNewConstMetric(c.metrics["chassis_health"].desc, prometheus.GaugeValue, chassisStatusHealthValue, ChassisLabelValues...)
			}
			if chassisStatusStateValue, ok := parseCommonStatusState(chassisStatusState); ok {
				ch <- prometheus.MustNewConstMetric(c.metrics["chassis_state"].desc, prometheus.GaugeValue, chassisStatusStateValue, ChassisLabelValues...)
			}

			if chassisThermal, err := chassis.Thermal(); err != nil || chassisThermal == nil {
				log.Infof("Errors Getting Thermal from chassis : %s", err)
			} else {
				// process temperature
				chassisTemperatures := chassisThermal.Temperatures
				wg := &sync.WaitGroup{}
				wg.Add(len(chassisTemperatures))

				for _, chassisTemperature := range chassisTemperatures {
					go parseChassisTemperature(ch, chassisID, chassisTemperature, wg)
				}

				// process fans

				chassisFans := chassisThermal.Fans
				wg2 := &sync.WaitGroup{}
				wg2.Add(len(chassisFans))
				for _, chassisFan := range chassisFans {
					go parseChassisFan(ch, chassisID, chassisFan, wg2)
				}
			}

			if chassisPowerInfo, err := chassis.Power(); err != nil || chassisPowerInfo == nil {
				log.Infof("Errors Getting powerinf from chassis : %s", err)
                                //os.Stderr.WriteString("cats")
			} else {
				//os.Stderr.WriteString("dogs")
				// power votages
				chassisPowerInfoVoltages := chassisPowerInfo.Voltages
				wg3 := &sync.WaitGroup{}
				wg3.Add(len(chassisPowerInfoVoltages))
				for _, chassisPowerInfoVoltage := range chassisPowerInfoVoltages {
					go parseChassisPowerInfoVoltage(ch, chassisID, chassisPowerInfoVoltage, wg3)

				}

				// powerSupply
				chassisPowerInfoPowerSupplies := chassisPowerInfo.PowerSupplies
				wg4 := &sync.WaitGroup{}
				wg4.Add(len(chassisPowerInfoPowerSupplies))
				for _, chassisPowerInfoPowerSupply := range chassisPowerInfoPowerSupplies {
					// os.Stderr.WriteString("birbs")
					go parseChassisPowerInfoPowerSupply(ch, chassisID, chassisPowerInfoPowerSupply, wg4)
				}

				chassisPowerControls := chassisPowerInfo.PowerControl
				wg99 := &sync.WaitGroup{}
				wg99.Add(len(chassisPowerControls))
				for _, chassisPowerControl := range chassisPowerControls {
					os.Stderr.WriteString("!")
					//os.Stderr.WriteString(chassisPowerControl)
					go parseChassisPowerControl(ch, chassisID, chassisPowerControl, wg99)
				}

			}

			// process NetapAdapter

			if networkAdapters, err := chassis.NetworkAdapters(); err != nil || networkAdapters == nil{
				log.Infof("Errors Getting NetworkAdapters from chassis : %s", err)
			} else {
				wg5 := &sync.WaitGroup{}
				wg5.Add(len(networkAdapters))

				for _, networkAdapter := range networkAdapters {

					go parseNetworkAdapter(ch, chassisID, networkAdapter, wg5)

				}
			}

		}
	}
	c.collectorScrapeStatus.WithLabelValues("chassis").Set(float64(1))
}

func parseChassisTemperature(ch chan<- prometheus.Metric, chassisID string, chassisTemperature redfish.Temperature, wg *sync.WaitGroup) {
	defer wg.Done()
	chassisTemperatureSensorName := chassisTemperature.Name
	//chassisTemperatureSensorName := chassisTemperature.ID
	chassisTemperatureSensorID := chassisTemperature.MemberID
	chassisTemperatureStatus := chassisTemperature.Status
	//			chassisTemperatureStatusHealth :=chassisTemperatureStatus.Health
	chassisTemperatureStatusState := chassisTemperatureStatus.State
	//			chassisTemperatureStatusLabelNames :=[]string{BaseLabelNames,"temperature_sensor_name","temperature_sensor_member_id")
	chassisTemperatureLabelvalues := []string{"temperature", chassisID, chassisTemperatureSensorName, chassisTemperatureSensorID}

	//		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_temperature_status_health"].desc, prometheus.GaugeValue, parseCommonStatusHealth(chassisTemperatureStatusHealth), chassisTemperatureLabelvalues...)
	if chassisTemperatureStatusStateValue, ok := parseCommonStatusState(chassisTemperatureStatusState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_temperature_sensor_state"].desc, prometheus.GaugeValue, chassisTemperatureStatusStateValue, chassisTemperatureLabelvalues...)
	}

	chassisTemperatureReadingCelsius := chassisTemperature.ReadingCelsius
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_temperature_celsius"].desc, prometheus.GaugeValue, float64(chassisTemperatureReadingCelsius), chassisTemperatureLabelvalues...)
}
func parseChassisFan(ch chan<- prometheus.Metric, chassisID string, chassisFan redfish.Fan, wg *sync.WaitGroup) {
	defer wg.Done()
	//chassisFanID := chassisFan.ID
        chassisFanID := chassisFan.MemberID
	chassisFanName := chassisFan.Name
	//chassisFanName := chassisFan.ID
	chassisFanStaus := chassisFan.Status
	chassisFanStausHealth := chassisFanStaus.Health
	chassisFanStausState := chassisFanStaus.State
	chassisFanRPM := chassisFan.Reading

	//			chassisFanStatusLabelNames :=[]string{BaseLabelNames,"fan_name","fan_member_id")
	chassisFanLabelvalues := []string{"fan", chassisID, chassisFanName, chassisFanID}

	if chassisFanStausHealthValue, ok := parseCommonStatusHealth(chassisFanStausHealth); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_health"].desc, prometheus.GaugeValue, chassisFanStausHealthValue, chassisFanLabelvalues...)
	}

	if chassisFanStausStateValue, ok := parseCommonStatusState(chassisFanStausState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_state"].desc, prometheus.GaugeValue, chassisFanStausStateValue, chassisFanLabelvalues...)
	}
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_fan_rpm"].desc, prometheus.GaugeValue, float64(chassisFanRPM), chassisFanLabelvalues...)

}

func parseChassisPowerInfoVoltage(ch chan<- prometheus.Metric, chassisID string, chassisPowerInfoVoltage redfish.Voltage, wg *sync.WaitGroup) {
	defer wg.Done()
	chassisPowerInfoVoltageName := chassisPowerInfoVoltage.Name
	chassisPowerInfoVoltageID := chassisPowerInfoVoltage.MemberID
	//chassisPowerInfoVoltageID := chassisPowerInfoVoltage.ID
	chassisPowerInfoVoltageNameReadingVolts := chassisPowerInfoVoltage.ReadingVolts
	chassisPowerInfoVoltageState := chassisPowerInfoVoltage.Status.State
	chassisPowerVotageLabelvalues := []string{"power_votage", chassisID, chassisPowerInfoVoltageName, chassisPowerInfoVoltageID}
	if chassisPowerInfoVoltageStateValue, ok := parseCommonStatusState(chassisPowerInfoVoltageState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_voltage_state"].desc, prometheus.GaugeValue, chassisPowerInfoVoltageStateValue, chassisPowerVotageLabelvalues...)
	}
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_voltage_volts"].desc, prometheus.GaugeValue, float64(chassisPowerInfoVoltageNameReadingVolts), chassisPowerVotageLabelvalues...)
}

func parseChassisPowerInfoPowerSupply(ch chan<- prometheus.Metric, chassisID string, chassisPowerInfoPowerSupply redfish.PowerSupply, wg *sync.WaitGroup) {

	defer wg.Done()
	chassisPowerInfoPowerSupplyName := chassisPowerInfoPowerSupply.Name
	chassisPowerInfoPowerSupplyID := chassisPowerInfoPowerSupply.ID
	chassisPowerInfoPowerSupplyPowerCapacityWatts := chassisPowerInfoPowerSupply.PowerCapacityWatts
	chassisPowerInfoPowerSupplyLastPowerOutputWatts := chassisPowerInfoPowerSupply.LastPowerOutputWatts

	//s := fmt.Sprintf("%f",chassisPowerInfoPowerSupply.LastPowerOutputWatts)
	//os.Stderr.WriteString("[ "+s+"W ]"+chassisID) 

	chassisPowerInfoPowerSupplyState := chassisPowerInfoPowerSupply.Status.State
	chassisPowerInfoPowerSupplyHealth := chassisPowerInfoPowerSupply.Status.Health
	chassisPowerSupplyLabelvalues := []string{"power_supply", chassisID, chassisPowerInfoPowerSupplyName, chassisPowerInfoPowerSupplyID}
	if chassisPowerInfoPowerSupplyStateValue, ok := parseCommonStatusState(chassisPowerInfoPowerSupplyState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_state"].desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyStateValue, chassisPowerSupplyLabelvalues...)
	}
	if chassisPowerInfoPowerSupplyHealthValue, ok := parseCommonStatusHealth(chassisPowerInfoPowerSupplyHealth); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_health"].desc, prometheus.GaugeValue, chassisPowerInfoPowerSupplyHealthValue, chassisPowerSupplyLabelvalues...)
	}
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_last_power_output_watts"].desc, prometheus.GaugeValue, float64(chassisPowerInfoPowerSupplyLastPowerOutputWatts), chassisPowerSupplyLabelvalues...)
	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powersupply_power_capacity_watts"].desc, prometheus.GaugeValue, float64(chassisPowerInfoPowerSupplyPowerCapacityWatts), chassisPowerSupplyLabelvalues...)
}

func parseChassisPowerControl(ch chan<- prometheus.Metric, chassisID string, chassisPowerControl redfish.PowerControl, wg *sync.WaitGroup) {
	
	defer wg.Done()
	os.Stderr.WriteString(chassisID)

	s := fmt.Sprintf("%f",chassisPowerControl.PowerConsumedWatts )
        os.Stderr.WriteString("Consumed Watts: "+s)
	
	paw := fmt.Sprintf("%f",chassisPowerControl.PowerAllocatedWatts) 
	os.Stderr.WriteString("Allocated Watts: "+paw)

	pa := fmt.Sprintf("%f",chassisPowerControl.PowerAvailableWatts )
	os.Stderr.WriteString("PowerAvailableWatts : "+pa)

	pcw := fmt.Sprintf("%f",chassisPowerControl.PowerCapacityWatts )
	os.Stderr.WriteString("PowerCapacityWatts  : "+pcw)

	//s := fmt.Sprintf("%f",chassisPowerControl.PowerLimit)
	//s.Stderr.WriteString("PowerCapacityWatts  : "+s)

	//s := fmt.Sprintf("%f",chassisPowerControl.PowerLimit)
	//s.Stderr.WriteString("PowerCapacityWatts  : "+s)

	chassisPowerControlLabelvalues := []string{"power_supply", chassisID, "", ""}

	ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_power_powercontrol_powerallocatedwatts"].desc, prometheus.GaugeValue, float64(chassisPowerControl.PowerAllocatedWatts), chassisPowerControlLabelvalues...)

}


func parseNetworkAdapter(ch chan<- prometheus.Metric, chassisID string, networkAdapter *redfish.NetworkAdapter, wg *sync.WaitGroup) {

	defer wg.Done()
	networkAdapterName := networkAdapter.Name
	networkAdapterID := networkAdapter.ID
	networkAdapterState := networkAdapter.Status.State
	networkAdapterHealthState := networkAdapter.Status.Health
	chassisNetworkAdapterLabelValues := []string{"network_adapter", chassisID, networkAdapterName, networkAdapterID}
	if networkAdapterStateValue, ok := parseCommonStatusState(networkAdapterState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_network_adapter_state"].desc, prometheus.GaugeValue, networkAdapterStateValue, chassisNetworkAdapterLabelValues...)
	}
	if networkAdapterHealthStateValue, ok := parseCommonStatusHealth(networkAdapterHealthState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_network_adapter_health_state"].desc, prometheus.GaugeValue, networkAdapterHealthStateValue, chassisNetworkAdapterLabelValues...)
	}

	if networkPorts, err := networkAdapter.NetworkPorts(); err != nil {
		log.Infof("Errors Getting Network port from networkAdapter : %s", err)
	} else {
		wg6 := &sync.WaitGroup{}
		wg6.Add(len(networkPorts))
		for _, networkPort := range networkPorts {
			go parseNetworkPort(ch, chassisID, networkPort, networkAdapterName,networkAdapterID, wg6)
		}

	}
}

func parseNetworkPort(ch chan<- prometheus.Metric, chassisID string, networkPort *redfish.NetworkPort, networkAdapterName string, networkAdapterID string, wg *sync.WaitGroup) {
	defer wg.Done()
	networkPortName := networkPort.Name
	networkPortID := networkPort.ID
	networkPortState := networkPort.Status.State
	networkPortLinkType := networkPort.ActiveLinkTechnology
	networkPortLinkSpeed := fmt.Sprintf("%d Mbps", networkPort.CurrentLinkSpeedMbps)
	networkPortHealthState := networkPort.Status.Health
	chassisNetworkPortLabelValues := []string{"network_port", chassisID, networkAdapterName, networkAdapterID, networkPortName, networkPortID, string(networkPortLinkType), networkPortLinkSpeed}
	if networkPortStateValue, ok := parseCommonStatusState(networkPortState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_network_port_state"].desc, prometheus.GaugeValue, networkPortStateValue, chassisNetworkPortLabelValues...)
	}
	if networkPortHealthStateValue, ok := parseCommonStatusHealth(networkPortHealthState); ok {
		ch <- prometheus.MustNewConstMetric(chassisMetrics["chassis_network_port_health_state"].desc, prometheus.GaugeValue, networkPortHealthStateValue, chassisNetworkPortLabelValues...)
	}
}
