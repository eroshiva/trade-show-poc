-- Create "network_device_child_network_devices" table
CREATE TABLE "network_device_child_network_devices" (
  "network_device_id" character varying NOT NULL,
  "parent_network_device_id" character varying NOT NULL,
  PRIMARY KEY ("network_device_id", "parent_network_device_id"),
  CONSTRAINT "network_device_child_network_devices_network_device_id" FOREIGN KEY ("network_device_id") REFERENCES "network_devices" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "network_device_child_network_devices_parent_network_device_id" FOREIGN KEY ("parent_network_device_id") REFERENCES "network_devices" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
