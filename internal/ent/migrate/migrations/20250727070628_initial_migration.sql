-- Create "versions" table
CREATE TABLE "versions" (
  "id" character varying NOT NULL,
  "version" character varying NOT NULL,
  "checksum" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "network_devices" table
CREATE TABLE "network_devices" (
  "id" character varying NOT NULL,
  "vendor" character varying NOT NULL,
  "model" character varying NOT NULL,
  "hw_version" character varying NULL,
  "network_device_sw_version" character varying NULL,
  "network_device_fw_version" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "network_devices_versions_fw_version" FOREIGN KEY ("network_device_fw_version") REFERENCES "versions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "network_devices_versions_sw_version" FOREIGN KEY ("network_device_sw_version") REFERENCES "versions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "device_status" table
CREATE TABLE "device_status" (
  "id" character varying NOT NULL,
  "status" character varying NOT NULL,
  "last_seen" character varying NULL,
  "device_status_network_device" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "device_status_network_devices_network_device" FOREIGN KEY ("device_status_network_device") REFERENCES "network_devices" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "endpoints" table
CREATE TABLE "endpoints" (
  "id" character varying NOT NULL,
  "host" character varying NOT NULL,
  "port" character varying NOT NULL,
  "protocol" character varying NOT NULL,
  "network_device_endpoints" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "endpoints_network_devices_endpoints" FOREIGN KEY ("network_device_endpoints") REFERENCES "network_devices" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
