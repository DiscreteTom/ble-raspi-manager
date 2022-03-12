device = await navigator.bluetooth.requestDevice({
  acceptAllDevices: true,
  optionalServices: ["f846752b-af47-43ed-bdf0-fba82da6fd58"],
});
server = await device.gatt.connect();
service = await server.getPrimaryService(
  "f846752b-af47-43ed-bdf0-fba82da6fd58"
);
char = await service.getCharacteristic("e94f5099-db86-4b29-a4ce-08033fda1a7d");
value = await char.readValue();
await char.writeValue(new Uint8Array([0, 1, 2, 3, 4]));
