import { computed, nextTick, ref, type Ref } from "vue";
import Sortable from "sortablejs";


export default abstract class SmartTable {
  dataList = ref([]);
  editMap = ref({});
  activeX = ref(-1);
  loading: Ref<boolean> = ref(false);

  abstract submitSave(dict: any): Promise<Result>;
  abstract loadData(): void;

  editingData = computed(() => {
    return (index: number) => {
      return this.editMap.value[index]?.editing;
    };
  });

  onMouseleave(index: number) {
    this.editingData.value[index]
      ? (this.activeX.value = index)
      : (this.activeX.value = -1);
  }

  onMouseenter(index: number) {
    this.activeX.value = index;
  }

  onEdit(row: any, index: number) {
    this.editMap.value[index] = Object.assign({ ...row, editing: true });
  }

  onSave(index: number) {
    this.submitSave(this.dataList.value[index]).then(() => {
      this.editMap.value[index].editing = false;
      this.loadData();
    });
  }

  isLoading() {
    return this.loading.value;
  }

  rowDrop = (event: { preventDefault: () => void }, cssIdSelector: string) => {
    event.preventDefault();

    nextTick(() => {
      const wrapper: HTMLElement = document.querySelector(
        cssIdSelector + " .el-table__body-wrapper tbody"
      );

      Sortable.create(wrapper, {
        animation: 300,
        handle: ".drag-btn",
        onEnd: ({ newIndex, oldIndex }) => {
          const currentRow = this.dataList.value.splice(oldIndex, 1)[0];
          this.dataList.value.splice(newIndex, 0, currentRow);
          this.dataList.value.forEach((item, index) => {
            item.sort = index + 1 + "";
          });
        }
      });
    }).then(() => {});
  };

  editable = (row: any, index: number, field: string) => (
    <div
      class="flex-bc w-full h-[32px]"
      onMouseenter={() => this.onMouseenter(index)}
      onMouseleave={() => this.onMouseleave(index)}
    >
      {!this.editingData.value(index) ? (
        <div
          style={{ width: "calc(100%)" }}
          ondblclick={() => this.onEdit(row, index)}
        >
          {row[field]}
        </div>
      ) : (
        <>
          <el-input
            v-model={row[field]}
            onblur={e => {
              console.log("Blur triggered!", e);
              console.trace("Blur call stack:");
            }}
            onkeydown={e => {
              if (e.key === "Enter") {
                this.onSave(index);
              }
            }}
          />
        </>
      )}
    </div>
  );
  changeStatus = (row: any, index: number) => (
    <>
      {this.editMap.value[index]?.editing ? (
        <el-switch
          v-model={row.status}
          inline-prompt
          active-value={"1"}
          inactive-value={"2"}
          active-text="启用"
          inactive-text="禁用"
        />
      ) : (
        <el-tag type={row.status === "1" ? "primary" : "danger"}>
          {row.status === "1" ? "启用" : "禁用"}
        </el-tag>
      )}
    </>
  );

  sortable = (cssIdSelector: string) => (
    <div class="flex items-center">
      <iconify-icon-online
        icon="icon-park-outline:drag"
        class="drag-btn cursor-grab"
        onMouseenter={(event: { preventDefault: () => void }) =>
          this.rowDrop(event, cssIdSelector)
        }
      />
    </div>
  );

  addRow(data: any) {
    this.dataList.value.push(data);
    this.editMap.value[this.dataList.value.length - 1] = data;
  }

  removeRow(index: number) {
    this.dataList.value.splice(index, 1);
  }
}
